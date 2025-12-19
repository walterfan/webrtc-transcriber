package transcribe

import (
	"io"
)

// Result is the struct used to serialize the results back to the client
type Result struct {
	Text       string  `json:"text"`
	Confidence float32 `json:"confidence"`
	Final      bool    `json:"final"`
	AudioFile  string  `json:"audio_file,omitempty"`
	TextFile   string  `json:"text_file,omitempty"`
}

// StreamOptions contains options for creating a transcription stream
type StreamOptions struct {
	Language   string // Language code (e.g., "en", "zh", "auto")
	Transcribe bool   // Whether to transcribe (if false, just record)
}

// Service is an abstract representation of the transcription service
type Service interface {
	CreateStream() (Stream, error)
	CreateStreamWithOptions(opts StreamOptions) (Stream, error)
}

// Stream is an abstract representation of a transcription stream
type Stream interface {
	io.Writer
	io.Closer
	Results() <-chan Result
}
