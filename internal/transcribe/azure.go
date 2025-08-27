package transcribe

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"

	"github.com/gorilla/websocket"
)

// AzureTranscriber is the implementation of the transcribe.Service,
// using Microsoft Azure Speech Service for speech recognition
type AzureTranscriber struct {
	subscriptionKey string
	region          string
	ctx             context.Context
}

// AzureStream implements the transcribe.Stream interface,
// it handles the WebSocket connection to Azure Speech Service
type AzureStream struct {
	conn    *websocket.Conn
	results chan Result
	ctx     context.Context
}

// Azure Speech Service message structures
type azureSpeechConfig struct {
	System struct {
		Name    string `json:"name"`
		Version string `json:"version"`
	} `json:"system"`
}

type azureSpeechRequest struct {
	Context azureSpeechConfig `json:"context"`
	Audio   struct {
		ContentType string `json:"contentType"`
		Data        string `json:"data"`
	} `json:"audio"`
}

type azureSpeechResponse struct {
	Type        string `json:"type"`
	ID          string `json:"id"`
	Timestamp   string `json:"timestamp"`
	Recognition struct {
		DisplayText string  `json:"displayText"`
		Offset      int64   `json:"offset"`
		Duration    int64   `json:"duration"`
		Confidence  float64 `json:"confidence"`
	} `json:"recognition"`
	Status string `json:"status"`
}

// CreateStream creates a new transcription stream
func (a *AzureTranscriber) CreateStream() (Stream, error) {
	// Generate WebSocket URL for Azure Speech Service
	wsURL := fmt.Sprintf("wss://%s.stt.speech.microsoft.com/speech/recognition/conversation/cognitiveservices/v1?api-version=2021-08-01-preview", a.region)

	// Create WebSocket connection
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, map[string][]string{
		"Ocp-Apim-Subscription-Key": {a.subscriptionKey},
		"Content-Type":              {"application/json"},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Azure Speech Service: %w", err)
	}

	// Send initial configuration
	config := azureSpeechConfig{}
	config.System.Name = "webrtc-transcriber"
	config.System.Version = "1.0.0"

	configMsg := map[string]interface{}{
		"type":    "speech.config",
		"context": config,
	}

	configBytes, err := json.Marshal(configMsg)
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := conn.WriteMessage(websocket.TextMessage, configBytes); err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to send config: %w", err)
	}

	stream := &AzureStream{
		conn:    conn,
		results: make(chan Result, 10),
		ctx:     a.ctx,
	}

	// Start listening for responses
	go stream.listenForResults()

	return stream, nil
}

// Results returns a channel that will receive the transcription results
func (as *AzureStream) Results() <-chan Result {
	return as.results
}

// Close sends an end-of-stream marker and closes the WebSocket connection
func (as *AzureStream) Close() error {
	// Send end-of-stream marker
	endMsg := map[string]interface{}{
		"type": "audio.end",
	}

	endBytes, err := json.Marshal(endMsg)
	if err != nil {
		log.Printf("Warning: failed to marshal end message: %v", err)
	} else {
		if err := as.conn.WriteMessage(websocket.TextMessage, endBytes); err != nil {
			log.Printf("Warning: failed to send end message: %v", err)
		}
	}

	// Close WebSocket connection
	if err := as.conn.Close(); err != nil {
		log.Printf("Warning: failed to close WebSocket: %v", err)
	}

	// Close results channel
	close(as.results)

	return nil
}

// Write sends audio data to the Azure Speech Service
func (as *AzureStream) Write(buffer []byte) (int, error) {
	// Encode audio data as base64
	audioData := base64.StdEncoding.EncodeToString(buffer)

	// Create speech request
	request := azureSpeechRequest{}
	request.Context.System.Name = "webrtc-transcriber"
	request.Context.System.Version = "1.0.0"
	request.Audio.ContentType = "audio/wav;codecs=audio/pcm;rate=48000"
	request.Audio.Data = audioData

	// Marshal request
	requestBytes, err := json.Marshal(request)
	if err != nil {
		return 0, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Send audio data
	if err := as.conn.WriteMessage(websocket.TextMessage, requestBytes); err != nil {
		return 0, fmt.Errorf("failed to send audio data: %w", err)
	}

	return len(buffer), nil
}

// listenForResults listens for WebSocket messages and processes transcription results
func (as *AzureStream) listenForResults() {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Recovered from panic in Azure stream listener: %v", r)
		}
	}()

	for {
		select {
		case <-as.ctx.Done():
			return
		default:
			// Read message
			_, message, err := as.conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					log.Printf("WebSocket error: %v", err)
				}
				return
			}

			// Parse response
			var response azureSpeechResponse
			if err := json.Unmarshal(message, &response); err != nil {
				log.Printf("Failed to unmarshal response: %v", err)
				continue
			}

			// Process different response types
			switch response.Type {
			case "recognition":
				if response.Recognition.DisplayText != "" {
					// Send result
					result := Result{
						Text:       response.Recognition.DisplayText,
						Confidence: float32(response.Recognition.Confidence),
						Final:      response.Status == "success",
					}

					select {
					case as.results <- result:
						// Result sent successfully
					case <-as.ctx.Done():
						return
					default:
						// Channel is full, skip this result
						log.Printf("Results channel is full, skipping result")
					}
				}

			case "error":
				log.Printf("Azure Speech Service error: %s", response.Status)

			case "end":
				log.Printf("Azure Speech Service stream ended")
				return
			}
		}
	}
}

// NewAzureTranscriber creates a new instance of the transcribe.Service that uses Azure Speech Service
func NewAzureTranscriber(ctx context.Context, subscriptionKey, region string) (Service, error) {
	if subscriptionKey == "" || region == "" {
		return nil, fmt.Errorf("subscriptionKey and region are required")
	}

	return &AzureTranscriber{
		subscriptionKey: subscriptionKey,
		region:          region,
		ctx:             ctx,
	}, nil
}
