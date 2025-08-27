package rtc

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/pion/webrtc/v2"
	"github.com/walterfan/webrtc-transcriber/internal/transcribe"
)

// PionPeerConnection is a webrtc.PeerConnection wrapper that implements the
// PeerConnection interface
type PionPeerConnection struct {
	pc *webrtc.PeerConnection
}

// PionRtcService is our implementation of the rtc.Service
type PionRtcService struct {
	stunServer  string
	transcriber transcribe.Service
}

// NewPionRtcService creates a new instances of PionRtcService
func NewPionRtcService(stun string, transcriber transcribe.Service) Service {
	return &PionRtcService{
		stunServer:  stun,
		transcriber: transcriber,
	}
}

// ProcessOffer handles the SDP offer coming from the client,
// return the SDP answer that must be passed back to stablish the WebRTC
// connection.
func (p *PionPeerConnection) ProcessOffer(offer string) (string, error) {
	err := p.pc.SetRemoteDescription(webrtc.SessionDescription{
		SDP:  offer,
		Type: webrtc.SDPTypeOffer,
	})
	if err != nil {
		return "", err
	}

	answer, err := p.pc.CreateAnswer(nil)
	if err != nil {
		return "", err
	}
	err = p.pc.SetLocalDescription(answer)
	if err != nil {
		return "", err
	}
	return answer.SDP, nil
}

// Close just closes the underlying peer connection
func (p *PionPeerConnection) Close() error {
	return p.pc.Close()
}

func (pi *PionRtcService) handleAudioTrack(track *webrtc.Track, dc *webrtc.DataChannel) error {
	// Safety check for nil parameters
	if track == nil {
		return fmt.Errorf("track is nil")
	}
	if dc == nil {
		return fmt.Errorf("dataChannel is nil")
	}
	if pi.transcriber == nil {
		return fmt.Errorf("transcriber service is nil")
	}

	decoder, err := newDecoder()
	if err != nil {
		return err
	}
	trStream, err := pi.transcriber.CreateStream()
	if err != nil {
		return err
	}
	defer func() {
		err := trStream.Close()
		if err != nil {
			log.Printf("Error closing stream %v", err)
			return
		}
		for result := range trStream.Results() {
			log.Printf("Result: %v", result)
			msg, err := json.Marshal(result)
			if err != nil {
				continue
			}
			err = dc.Send(msg)
			if err != nil {
				fmt.Printf("DataChannel error: %v", err)
			}
		}
		dc.Close()
	}()

	errs := make(chan error, 2)
	audioStream := make(chan []byte, 100)   // Buffered channel to avoid blocking
	response := make(chan bool, 100)        // Buffered channel to avoid blocking
	timer := time.NewTimer(5 * time.Second) // 5 second timeout for normal operation
	defer timer.Stop()

	// Context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		defer close(audioStream)
		for {
			select {
			case <-ctx.Done():
				return
			default:
				packet, err := track.ReadRTP()
				if err != nil {
					if err == io.EOF {
						log.Printf("Track ended for %s", track.ID())
						return
					}
					log.Printf("Error reading RTP packet: %v", err)
					errs <- err
					return
				}

				// Reset timer on successful read
				timer.Reset(5 * time.Second)

				select {
				case audioStream <- packet.Payload:
					// Wait for response before continuing
					select {
					case <-response:
						// Continue reading
					case <-ctx.Done():
						return
					}
				case <-ctx.Done():
					return
				}
			}
		}
	}()

	err = nil
	for {
		select {
		case audioChunk, ok := <-audioStream:
			if !ok {
				// Channel closed, stream ended
				log.Printf("Audio stream ended for track %s", track.ID())
				return nil
			}

			payload, err := decoder.decode(audioChunk)
			if err != nil {
				log.Printf("Error decoding audio: %v", err)
				continue // Skip this chunk but continue processing
			}

			// Send response to unblock the reader
			select {
			case response <- true:
			default:
				// Response channel is full, skip
			}

			_, err = trStream.Write(payload)
			if err != nil {
				log.Printf("Error writing to transcriber: %v", err)
				return err
			}

		case <-timer.C:
			log.Printf("Read operation timed out for track %s, closing stream", track.ID())
			cancel() // Signal shutdown
			return nil

		case err = <-errs:
			log.Printf("Unexpected error reading track %s: %v", track.ID(), err)
			cancel() // Signal shutdown
			return err

		case <-ctx.Done():
			log.Printf("Context cancelled for track %s", track.ID())
			return nil
		}
	}
}

// CreatePeerConnection creates and configures a new peer connection for
// our purposes, receive one audio track and send data through one DataChannel
func (pi *PionRtcService) CreatePeerConnection() (PeerConnection, error) {
	pcconf := webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			webrtc.ICEServer{
				URLs: []string{pi.stunServer},
			},
		},
		SDPSemantics: webrtc.SDPSemanticsUnifiedPlanWithFallback,
	}
	pc, err := webrtc.NewPeerConnection(pcconf)
	if err != nil {
		return nil, err
	}

	// Use a buffered channel to avoid blocking
	dataChan := make(chan *webrtc.DataChannel, 1)
	var audioTrack *webrtc.Track
	var dataChannel *webrtc.DataChannel

	// Helper function to start audio processing when both are ready
	startAudioProcessing := func() {
		if audioTrack != nil && dataChannel != nil {
			log.Printf("Starting audio processing for track %s with DataChannel %s", audioTrack.ID(), dataChannel.Label())
			go func() {
				err := pi.handleAudioTrack(audioTrack, dataChannel)
				if err != nil {
					log.Printf("Error reading track (%s): %v\n", audioTrack.ID(), err)
				}
			}()
		} else {
			log.Printf("Not ready to start audio processing: audioTrack=%v, dataChannel=%v",
				audioTrack != nil, dataChannel != nil)
		}
	}

	pc.OnDataChannel(func(dc *webrtc.DataChannel) {
		log.Printf("DataChannel established: %s", dc.Label())
		dataChannel = dc
		select {
		case dataChan <- dc:
		default:
			// Channel is full, replace the value
		}
		// Only start audio processing if we have both components
		if audioTrack != nil && dataChannel != nil {
			startAudioProcessing()
		}
	})

	pc.OnTrack(func(track *webrtc.Track, r *webrtc.RTPReceiver) {
		if track.Codec().Name == "opus" {
			//log.Printf("Received audio (%s) track, id = %s\n", track.Codec().Name, track.ID())
			audioTrack = track
			// Only start audio processing if we have both components
			if audioTrack != nil && dataChannel != nil {
				startAudioProcessing()
			}
		}
	})

	pc.OnICEConnectionStateChange(func(connState webrtc.ICEConnectionState) {
		log.Printf("Connection state: %s \n", connState.String())
	})

	_, err = pc.AddTransceiver(webrtc.RTPCodecTypeAudio, webrtc.RtpTransceiverInit{
		Direction: webrtc.RTPTransceiverDirectionRecvonly,
	})
	if err != nil {
		log.Printf("Can't add transceiver: %s", err)
		return nil, err
	}

	return &PionPeerConnection{
		pc: pc,
	}, nil
}
