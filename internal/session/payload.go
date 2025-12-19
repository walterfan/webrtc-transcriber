package session

type newSessionRequest struct {
	Offer      string `json:"offer"`
	Language   string `json:"language,omitempty"`   // Language code for transcription (e.g., "en", "zh", "auto")
	Transcribe *bool  `json:"transcribe,omitempty"` // Whether to transcribe (default: true)
}

type newSessionResponse struct {
	Answer string `json:"answer"`
}
