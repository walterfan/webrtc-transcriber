package rtc

import (
	"io"
)

// PeerConnectionOptions contains options for creating a peer connection
type PeerConnectionOptions struct {
	Language   string // Language code for transcription (e.g., "en", "zh", "auto")
	Transcribe bool   // Whether to transcribe audio (default: true)
}

// PeerConnection Represents a WebRTC connection to a single peer
type PeerConnection interface {
	io.Closer
	ProcessOffer(offer string) (string, error)
}

// Service WebRTC service
type Service interface {
	CreatePeerConnection() (PeerConnection, error)
	CreatePeerConnectionWithOptions(opts PeerConnectionOptions) (PeerConnection, error)
}
