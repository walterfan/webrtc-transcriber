package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sort"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/walterfan/webrtc-transcriber/internal/rtc"
	"github.com/walterfan/webrtc-transcriber/internal/session"
	"github.com/walterfan/webrtc-transcriber/internal/transcribe"
)

const (
	httpDefaultPort      = "9070"
	defaultStunServer    = "stun:stun.l.google.com:19302"
	defaultRecordingsDir = "recordings"
	sessionCookieName    = "session_token"
	sessionDuration      = 24 * time.Hour
)

// Session management
type SessionStore struct {
	sessions map[string]SessionData
	mu       sync.RWMutex
}

type SessionData struct {
	Username  string
	ExpiresAt time.Time
}

var sessionStore = &SessionStore{
	sessions: make(map[string]SessionData),
}

// accounts stores username:password pairs loaded from environment
var accounts = make(map[string]string)

// loadAccounts parses the accounts from environment variable
// Format: "alice:abc, walter:abd"
func loadAccounts() {
	accountsEnv := os.Getenv("accounts")
	if accountsEnv == "" {
		log.Printf("Warning: No accounts configured in .env file (accounts=username:password,...)")
		return
	}

	pairs := strings.Split(accountsEnv, ",")
	for _, pair := range pairs {
		pair = strings.TrimSpace(pair)
		if pair == "" {
			continue
		}
		parts := strings.SplitN(pair, ":", 2)
		if len(parts) == 2 {
			username := strings.TrimSpace(parts[0])
			password := strings.TrimSpace(parts[1])
			accounts[username] = password
			log.Printf("Loaded account: %s", username)
		}
	}

	if len(accounts) == 0 {
		log.Printf("Warning: No valid accounts found in accounts environment variable")
	}
}

// generateSessionToken creates a random session token
func generateSessionToken() string {
	bytes := make([]byte, 32)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// createSession creates a new session for a user
func (s *SessionStore) createSession(username string) string {
	s.mu.Lock()
	defer s.mu.Unlock()

	token := generateSessionToken()
	s.sessions[token] = SessionData{
		Username:  username,
		ExpiresAt: time.Now().Add(sessionDuration),
	}
	return token
}

// validateSession checks if a session token is valid
func (s *SessionStore) validateSession(token string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	session, exists := s.sessions[token]
	if !exists {
		return "", false
	}
	if time.Now().After(session.ExpiresAt) {
		return "", false
	}
	return session.Username, true
}

// deleteSession removes a session
func (s *SessionStore) deleteSession(token string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.sessions, token)
}

// authMiddleware wraps handlers to require authentication
func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip auth for login endpoint and static assets
		if r.URL.Path == "/login" || r.URL.Path == "/auth/status" {
			next.ServeHTTP(w, r)
			return
		}

		cookie, err := r.Cookie(sessionCookieName)
		if err != nil || cookie.Value == "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		_, valid := sessionStore.validateSession(cookie.Value)
		if !valid {
			http.Error(w, "Session expired", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// loginHandler handles login requests
func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse form data
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	username := r.FormValue("username")
	password := r.FormValue("password")

	// Validate credentials
	expectedPassword, exists := accounts[username]
	if !exists || expectedPassword != password {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"success": false, "message": "Invalid username or password"}`))
		return
	}

	// Create session
	token := sessionStore.createSession(username)

	// Set cookie
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		MaxAge:   int(sessionDuration.Seconds()),
		SameSite: http.SameSiteStrictMode,
	})

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(fmt.Sprintf(`{"success": true, "username": "%s"}`, username)))
}

// logoutHandler handles logout requests
func logoutHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(sessionCookieName)
	if err == nil && cookie.Value != "" {
		sessionStore.deleteSession(cookie.Value)
	}

	// Clear cookie
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1,
	})

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"success": true}`))
}

// authStatusHandler returns the current authentication status
func authStatusHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(sessionCookieName)
	if err != nil || cookie.Value == "" {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"authenticated": false}`))
		return
	}

	username, valid := sessionStore.validateSession(cookie.Value)
	if !valid {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"authenticated": false}`))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(fmt.Sprintf(`{"authenticated": true, "username": "%s"}`, username)))
}

// selectVendor selects the appropriate transcription service based on command line arguments
// and available credentials. Command line arguments take precedence over environment variables.
//
// Priority Order (when --vendor is specified):
// 1. Command line --vendor flag (highest priority)
// 2. Google Speech (if --google.cred flag provided)
// 3. Environment variable based selection (fallback)
//
// Supported vendors: google, azure, baidu, xunfei, whisper, recorder
func selectVendor(ctx context.Context, googleCred, vendor, model, output, language string, keepWav, keepTxt bool) (transcribe.Service, error) {
	// If vendor is specified via command line, use it directly
	if vendor != "" {
		switch vendor {
		case "google":
			if googleCred == "" {
				return nil, fmt.Errorf("--vendor=google requires --google.cred flag")
			}
			tr, err := transcribe.NewGoogleSpeech(ctx, googleCred)
			if err != nil {
				return nil, fmt.Errorf("failed to create Google Speech service: %w", err)
			}
			log.Printf("Using Google Speech service (via --vendor flag)")
			return tr, nil

		case "azure":
			azureKey := os.Getenv("AZURE_SPEECH_KEY")
			azureRegion := os.Getenv("AZURE_SPEECH_REGION")
			if azureKey == "" || azureRegion == "" {
				return nil, fmt.Errorf("--vendor=azure requires AZURE_SPEECH_KEY and AZURE_SPEECH_REGION environment variables")
			}
			tr, err := transcribe.NewAzureTranscriber(ctx, azureKey, azureRegion)
			if err != nil {
				return nil, fmt.Errorf("failed to create Azure Speech service: %w", err)
			}
			log.Printf("Using Azure Speech service (via --vendor flag, region: %s)", azureRegion)
			return tr, nil

		case "baidu":
			baiduAppID := os.Getenv("BAIDU_APP_ID")
			baiduApiKey := os.Getenv("BAIDU_API_KEY")
			baiduSecretKey := os.Getenv("BAIDU_SECRET_KEY")
			if baiduAppID == "" || baiduApiKey == "" || baiduSecretKey == "" {
				return nil, fmt.Errorf("--vendor=baidu requires BAIDU_APP_ID, BAIDU_API_KEY, and BAIDU_SECRET_KEY environment variables")
			}
			tr, err := transcribe.NewBaiduTranscriber(ctx, baiduAppID, baiduApiKey, baiduSecretKey)
			if err != nil {
				return nil, fmt.Errorf("failed to create Baidu Speech service: %w", err)
			}
			log.Printf("Using Baidu Speech service (via --vendor flag)")
			return tr, nil

		case "xunfei":
			appID := os.Getenv("XUNFEI_APP_ID")
			apiKey := os.Getenv("XUNFEI_API_KEY")
			apiSecret := os.Getenv("XUNFEI_API_SECRET")
			appUrl := os.Getenv("XUNFEI_API_URL")
			if appID == "" || apiKey == "" || apiSecret == "" {
				return nil, fmt.Errorf("--vendor=xunfei requires XUNFEI_APP_ID, XUNFEI_API_KEY, and XUNFEI_API_SECRET environment variables")
			}
			tr, err := transcribe.NewIflyTekTranscriber(ctx, appID, apiKey, apiSecret, appUrl)
			if err != nil {
				return nil, fmt.Errorf("failed to create Xunfei service: %w", err)
			}
			log.Printf("Using Xunfei (IflyTek) service (via --vendor flag)")
			return tr, nil

		case "whisper":
			// Use command line arguments for Whisper
			whisperModelPath := model
			whisperPath := os.Getenv("WHISPER_PATH")
			outputDir := output
			if outputDir == "" {
				outputDir = "./recordings"
			}

			tr, err := transcribe.NewWhisperTranscriber(ctx, whisperModelPath, whisperPath, outputDir, language, keepWav, keepTxt)
			if err != nil {
				// If Whisper is not available, fall back to Recorder service
				log.Printf("Whisper service not available: %v", err)
				log.Printf("Falling back to Recorder service")
				recorderTr, recorderErr := transcribe.NewRecorderTranscriber(ctx, outputDir)
				if recorderErr != nil {
					return nil, fmt.Errorf("failed to create Whisper service: %w, and failed to fallback to Recorder: %w", err, recorderErr)
				}
				log.Printf("Using Recorder service (fallback from Whisper, output: %s)", outputDir)
				return recorderTr, nil
			}
			log.Printf("Using Whisper service (via --vendor flag, model: %s, language: %s, output: %s)", model, language, outputDir)
			return tr, nil

		case "recorder":
			outputDir := output
			if outputDir == "" {
				outputDir = "./recordings"
			}

			tr, err := transcribe.NewRecorderTranscriber(ctx, outputDir)
			if err != nil {
				return nil, fmt.Errorf("failed to create Recorder service: %w", err)
			}
			log.Printf("Using Recorder service (via --vendor flag, output: %s)", outputDir)
			return tr, nil

		default:
			return nil, fmt.Errorf("unsupported vendor: %s. Supported vendors: google, azure, baidu, xunfei, whisper, recorder", vendor)
		}
	}

	// Fallback to automatic selection based on environment variables
	// Check Google Speech first (highest priority)
	if googleCred != "" {
		tr, err := transcribe.NewGoogleSpeech(ctx, googleCred)
		if err != nil {
			return nil, fmt.Errorf("failed to create Google Speech service: %w", err)
		}
		log.Printf("Using Google Speech service")
		return tr, nil
	}

	// Check Azure Speech credentials
	azureKey := os.Getenv("AZURE_SPEECH_KEY")
	azureRegion := os.Getenv("AZURE_SPEECH_REGION")
	if azureKey != "" && azureRegion != "" {
		tr, err := transcribe.NewAzureTranscriber(ctx, azureKey, azureRegion)
		if err != nil {
			return nil, fmt.Errorf("failed to create Azure Speech service: %w", err)
		}
		log.Printf("Using Azure Speech service (region: %s)", azureRegion)
		return tr, nil
	}

	// Check Baidu Speech credentials
	baiduAppID := os.Getenv("BAIDU_APP_ID")
	baiduApiKey := os.Getenv("BAIDU_API_KEY")
	baiduSecretKey := os.Getenv("BAIDU_SECRET_KEY")
	if baiduAppID != "" && baiduApiKey != "" && baiduSecretKey != "" {
		tr, err := transcribe.NewBaiduTranscriber(ctx, baiduAppID, baiduApiKey, baiduSecretKey)
		if err != nil {
			return nil, fmt.Errorf("failed to create Baidu Speech service: %w", err)
		}
		log.Printf("Using Baidu Speech service")
		return tr, nil
	}

	// Check Xunfei credentials
	appID := os.Getenv("XUNFEI_APP_ID")
	apiKey := os.Getenv("XUNFEI_API_KEY")
	apiSecret := os.Getenv("XUNFEI_API_SECRET")
	appUrl := os.Getenv("XUNFEI_API_URL")
	if appID != "" && apiKey != "" && apiSecret != "" {
		tr, err := transcribe.NewIflyTekTranscriber(ctx, appID, apiKey, apiSecret, appUrl)
		if err != nil {
			return nil, fmt.Errorf("failed to create Xunfei service: %w", err)
		}
		log.Printf("Using Xunfei (IflyTek) service")
		return tr, nil
	}

	// Check if Whisper is available (try auto-detection even without env vars)
	whisperModelPath := os.Getenv("WHISPER_MODEL_PATH")
	whisperPath := os.Getenv("WHISPER_PATH")
	outputDir := output
	if outputDir == "" {
		outputDir = os.Getenv("OUTPUT_PATH")
		if outputDir == "" {
			currentDir, err := os.Getwd()
			if err != nil {
				return nil, fmt.Errorf("failed to get current working directory: %w", err)
			}
			outputDir = currentDir + "/" + defaultRecordingsDir
		}
	}

	// Try to create Whisper service (will auto-detect if env vars are empty)
	whisperTr, err := transcribe.NewWhisperTranscriber(ctx, whisperModelPath, whisperPath, outputDir, language, keepWav, keepTxt)
	if err == nil {
		// Whisper service created successfully
		modelPath := whisperModelPath
		execPath := whisperPath
		if modelPath == "" {
			modelPath = "auto-detected"
		}
		if execPath == "" {
			execPath = "auto-detected"
		}
		log.Printf("Using Whisper service (model: %s, executable: %s, language: %s)", modelPath, execPath, language)
		return whisperTr, nil
	}

	// If Whisper failed, log the error but continue to next service
	log.Printf("Whisper service not available: %v", err)

	// Use Recorder service as fallback (no credentials needed)
	recorderOutputDir := output
	if recorderOutputDir == "" {
		recorderOutputDir = os.Getenv("RECORDER_OUTPUT_DIR")
		if recorderOutputDir == "" {
			recorderOutputDir = defaultRecordingsDir
		}
	}

	tr, err := transcribe.NewRecorderTranscriber(ctx, recorderOutputDir)
	if err != nil {
		return nil, fmt.Errorf("failed to create Recorder service: %w", err)
	}
	log.Printf("Using Recorder service (output directory: %s)", outputDir)
	return tr, nil
}

func main() {

	// Load environment variables from .env file before parsing flags
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: Error loading .env file: %v", err)
	}

	// Load accounts from environment
	loadAccounts()

	httpPort := flag.String("http.port", httpDefaultPort, "HTTP listen port")
	stunServer := flag.String("stun.server", defaultStunServer, "STUN server URL (stun:)")

	// New command line arguments
	vendor := flag.String("vendor", "whisper", "Transcription vendor: google, azure, baidu, xunfei, whisper, recorder")
	model := flag.String("model", "small", "Whisper model: tiny, base, small, medium, large")
	output := flag.String("output", "recordings", "Output directory for WAV and TXT files")
	language := flag.String("language", "auto", "Source language (e.g., en, cn, auto)")

	// File retention flags
	keepWav := flag.Bool("keep_wav", true, "Keep generated WAV files (default: true)")
	keepTxt := flag.Bool("keep_txt", true, "Keep generated TXT files (default: true)")

	// Add usage information
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Transcription Server - Real-time speech-to-text using WebRTC\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  # Use Google Speech with credentials\n")
		fmt.Fprintf(os.Stderr, "  export GOOGLE_CREDENTIALS=/path/to/credentials.json\n")
		fmt.Fprintf(os.Stderr, "  %s --vendor=google\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  # Use Azure Speech Service\n")
		fmt.Fprintf(os.Stderr, "  %s --vendor=azure\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  # Use Whisper with custom model and output (default vendor)\n")
		fmt.Fprintf(os.Stderr, "  %s --model=base --output=./my_output\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  # Use Recorder to save WAV files\n")
		fmt.Fprintf(os.Stderr, "  %s --vendor=recorder --output=./recordings\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  # Keep generated files\n")
		fmt.Fprintf(os.Stderr, "  %s --keep_wav --keep_txt\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Environment Variables:\n")
		fmt.Fprintf(os.Stderr, "  Environment variables can be set directly or loaded from a .env file\n")
		fmt.Fprintf(os.Stderr, "  GOOGLE_CREDENTIALS                        - Google Speech credentials file path\n")
		fmt.Fprintf(os.Stderr, "  AZURE_SPEECH_KEY, AZURE_SPEECH_REGION     - Azure Speech Service credentials\n")
		fmt.Fprintf(os.Stderr, "  BAIDU_APP_ID, BAIDU_API_KEY, BAIDU_SECRET_KEY - Baidu Speech credentials\n")
		fmt.Fprintf(os.Stderr, "  XUNFEI_APP_ID, XUNFEI_API_KEY, XUNFEI_API_SECRET, XUNFEI_API_URL - Xunfei credentials and API URL\n")
		fmt.Fprintf(os.Stderr, "  WHISPER_PATH                              - Path to Whisper executable\n")
	}

	flag.Parse()

	var tr transcribe.Service
	var err error
	ctx := context.Background()

	// Select transcription vendor based on available credentials
	googleCred := os.Getenv("GOOGLE_CREDENTIALS")
	tr, err = selectVendor(ctx, googleCred, *vendor, *model, *output, *language, *keepWav, *keepTxt)
	if err != nil {
		log.Fatalf("Failed to create transcription service: %v", err)
	}

	webrtc := rtc.NewPionRtcService(*stunServer, tr)
	// webrtc = rtc.NewLoggingService(webrtc)

	// Create a new mux for all routes
	mux := http.NewServeMux()

	// Public routes (no auth required)
	mux.HandleFunc("/login", loginHandler)
	mux.HandleFunc("/logout", logoutHandler)
	mux.HandleFunc("/auth/status", authStatusHandler)

	// Serve static assets (login page needs these)
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./web/index.html")
	})
	mux.Handle("/static/", http.StripPrefix("/static", http.FileServer(http.Dir("./web"))))

	// Protected routes (auth required)
	mux.Handle("/session", authMiddleware(session.MakeHandler(webrtc)))
	mux.Handle("/recordings/", authMiddleware(http.StripPrefix("/recordings", http.FileServer(http.Dir(*output)))))

	// Endpoint to list files in the recordings directory (protected)
	mux.HandleFunc("/files", func(w http.ResponseWriter, r *http.Request) {
		// Check authentication
		cookie, err := r.Cookie(sessionCookieName)
		if err != nil || cookie.Value == "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		_, valid := sessionStore.validateSession(cookie.Value)
		if !valid {
			http.Error(w, "Session expired", http.StatusUnauthorized)
			return
		}

		files, err := os.ReadDir(*output)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Collect file info with modification time
		type fileInfo struct {
			Name    string
			ModTime int64
		}
		var fileInfoList []fileInfo
		for _, file := range files {
			if !file.IsDir() {
				info, err := file.Info()
				if err != nil {
					continue
				}
				fileInfoList = append(fileInfoList, fileInfo{
					Name:    file.Name(),
					ModTime: info.ModTime().UnixMilli(),
				})
			}
		}

		// Sort by modification time descending (newest first)
		sort.Slice(fileInfoList, func(i, j int) bool {
			return fileInfoList[i].ModTime > fileInfoList[j].ModTime
		})

		// Return JSON response with file info
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("["))
		for i, f := range fileInfoList {
			if i > 0 {
				w.Write([]byte(","))
			}
			w.Write([]byte(fmt.Sprintf(`{"name":"%s","modTime":%d}`, f.Name, f.ModTime)))
		}
		w.Write([]byte("]"))
	})

	// Endpoint to delete a file in the recordings directory (protected)
	mux.HandleFunc("/delete/", func(w http.ResponseWriter, r *http.Request) {
		// Check authentication
		cookie, err := r.Cookie(sessionCookieName)
		if err != nil || cookie.Value == "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		_, valid := sessionStore.validateSession(cookie.Value)
		if !valid {
			http.Error(w, "Session expired", http.StatusUnauthorized)
			return
		}

		// Only allow DELETE method
		if r.Method != http.MethodDelete {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Extract filename from URL path
		filename := strings.TrimPrefix(r.URL.Path, "/delete/")
		if filename == "" {
			http.Error(w, "Filename required", http.StatusBadRequest)
			return
		}

		// Sanitize filename to prevent directory traversal
		filename = strings.ReplaceAll(filename, "..", "")
		filename = strings.ReplaceAll(filename, "/", "")
		filename = strings.ReplaceAll(filename, "\\", "")

		// Build full path
		filePath := fmt.Sprintf("%s/%s", *output, filename)

		// Check if file exists
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"success": false, "message": "File not found"}`))
			return
		}

		// Delete the file
		if err := os.Remove(filePath); err != nil {
			log.Printf("Error deleting file %s: %v", filePath, err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"success": false, "message": "Failed to delete file"}`))
			return
		}

		log.Printf("Deleted file: %s", filePath)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"success": true}`))
	})

	errors := make(chan error, 2)
	go func() {
		log.Printf("Starting signaling server on port %s", *httpPort)
		errors <- http.ListenAndServe(fmt.Sprintf(":%s", *httpPort), mux)
	}()

	go func() {
		interrupt := make(chan os.Signal, 1)
		signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)
		errors <- fmt.Errorf("received %v signal", <-interrupt)
	}()

	err = <-errors
	log.Printf("%s, exiting.", err)
}
