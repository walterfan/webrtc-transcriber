package session

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/walterfan/webrtc-transcriber/internal/rtc"
)

// MakeHandler returns an HTTP handler for the session service
func MakeHandler(webrtcService rtc.Service) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		dec := json.NewDecoder(r.Body)
		req := newSessionRequest{}

		if err := dec.Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// Log the language selection
		language := req.Language
		if language == "" {
			language = "auto"
		}

		// Default transcribe to true if not specified
		transcribe := true
		if req.Transcribe != nil {
			transcribe = *req.Transcribe
		}
		log.Printf("Creating peer connection with language: %s, transcribe: %v", language, transcribe)

		// Create peer connection with options
		peer, err := webrtcService.CreatePeerConnectionWithOptions(rtc.PeerConnectionOptions{
			Language:   language,
			Transcribe: transcribe,
		})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		answer, err := peer.ProcessOffer(req.Offer)

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		payload, err := json.Marshal(newSessionResponse{
			Answer: answer,
		})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Write(payload)
	})
	return mux
}
