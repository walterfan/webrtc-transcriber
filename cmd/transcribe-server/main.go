package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
	"github.com/walterfan/webrtc-transcriber/internal/rtc"
	"github.com/walterfan/webrtc-transcriber/internal/session"
	"github.com/walterfan/webrtc-transcriber/internal/transcribe"
)

const (
	httpDefaultPort      = "9070"
	defaultStunServer    = "stun:stun.l.google.com:19302"
	defaultRecordingsDir = "recordings"
)

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

			tr, err := transcribe.NewWhisperTranscriber(ctx, whisperModelPath, whisperPath, outputDir, keepWav, keepTxt)
			if err != nil {
				return nil, fmt.Errorf("failed to create Whisper service: %w", err)
			}
			log.Printf("Using Whisper service (via --vendor flag, model: %s, output: %s)", model, outputDir)
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
	whisperTr, err := transcribe.NewWhisperTranscriber(ctx, whisperModelPath, whisperPath, outputDir, keepWav, keepTxt)
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
		log.Printf("Using Whisper service (model: %s, executable: %s)", modelPath, execPath)
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

	httpPort := flag.String("http.port", httpDefaultPort, "HTTP listen port")
	stunServer := flag.String("stun.server", defaultStunServer, "STUN server URL (stun:)")

	// New command line arguments
	vendor := flag.String("vendor", "whisper", "Transcription vendor: google, azure, baidu, xunfei, whisper, recorder")
	model := flag.String("model", "tiny", "Whisper model: tiny, base, small, medium, large")
	output := flag.String("output", "recordings", "Output directory for WAV and TXT files")
	language := flag.String("language", "auto", "Source language (e.g., en, cn, auto)")

	// File retention flags
	keepWav := flag.Bool("keep_wav", false, "Keep generated WAV files (default: false)")
	keepTxt := flag.Bool("keep_txt", false, "Keep generated TXT files (default: false)")

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

	// Endpoint to create a new speech to text session
	http.Handle("/session", session.MakeHandler(webrtc))

	// Serve static assets
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./web/index.html")
	})
	http.Handle("/static/", http.StripPrefix("/static", http.FileServer(http.Dir("./web"))))

	errors := make(chan error, 2)
	go func() {
		log.Printf("Starting signaling server on port %s", *httpPort)
		errors <- http.ListenAndServe(fmt.Sprintf(":%s", *httpPort), nil)
	}()

	go func() {
		interrupt := make(chan os.Signal, 1)
		signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)
		errors <- fmt.Errorf("received %v signal", <-interrupt)
	}()

	err = <-errors
	log.Printf("%s, exiting.", err)
}
