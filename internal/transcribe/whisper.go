package transcribe

import (
	"context"
	"encoding/binary"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// WhisperTranscriber is the implementation of the transcribe.Service,
// using OpenAI's Whisper model for local speech recognition
type WhisperTranscriber struct {
	modelPath   string
	whisperPath string
	tempDir     string
	language    string // Language code (e.g., "en", "zh", "auto")
	ctx         context.Context
	mu          sync.Mutex
	counter     int
	keepWav     bool
	keepTxt     bool
}

// WhisperStream implements the transcribe.Stream interface,
// it handles audio processing and transcription using Whisper
type WhisperStream struct {
	filePath    string
	file        *os.File // Store the file handle
	results     chan Result
	ctx         context.Context
	transcriber *WhisperTranscriber
	language    string // Per-stream language override
	transcribe  bool   // Whether to transcribe (if false, just record)
	mu          sync.Mutex
	isClosed    bool
}

// WhisperConfig holds configuration for Whisper model
type WhisperConfig struct {
	Model       string  `json:"model"`       // Model size: tiny, base, small, medium, large
	Language    string  `json:"language"`    // Language code (e.g., "en", "zh", "auto")
	Task        string  `json:"task"`        // Task type: "transcribe" or "translate"
	Temperature float64 `json:"temperature"` // Sampling temperature (0.0 to 1.0)
}

// CreateStream creates a new transcription stream with default language
func (w *WhisperTranscriber) CreateStream() (Stream, error) {
	return w.CreateStreamWithOptions(StreamOptions{Language: w.language, Transcribe: true})
}

// CreateStreamWithOptions creates a new transcription stream with specified options
func (w *WhisperTranscriber) CreateStreamWithOptions(opts StreamOptions) (Stream, error) {
	w.mu.Lock()
	w.counter++
	streamID := w.counter
	w.mu.Unlock()

	// Use provided language or fall back to transcriber default
	language := opts.Language
	if language == "" {
		language = w.language
	}

	// Default transcribe to true if not explicitly set
	transcribe := opts.Transcribe

	// Create temporary file for audio data
	fileName := fmt.Sprintf("whisper_audio_%d_%s.wav", streamID, time.Now().Format("20060102_150405"))
	filePath := filepath.Join(w.tempDir, fileName)

	// Create output directory if it doesn't exist
	if err := os.MkdirAll(w.tempDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create output directory: %w", err)
	}

	// Create WAV file with header
	file, err := os.Create(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create WAV file: %w", err)
	}

	// Write WAV header (will be updated later with correct sizes)
	header := wavHeader{
		ChunkID:       [4]byte{'R', 'I', 'F', 'F'},
		Format:        [4]byte{'W', 'A', 'V', 'E'},
		Subchunk1ID:   [4]byte{'f', 'm', 't', ' '},
		Subchunk1Size: 16,
		AudioFormat:   1, // PCM
		NumChannels:   1, // Mono
		SampleRate:    48000,
		BitsPerSample: 16,
		Subchunk2ID:   [4]byte{'d', 'a', 't', 'a'},
	}

	// Calculate derived values
	header.ByteRate = header.SampleRate * uint32(header.NumChannels) * uint32(header.BitsPerSample) / 8
	header.BlockAlign = header.NumChannels * header.BitsPerSample / 8

	// Write header manually to ensure correct byte order
	if _, err := file.Write(header.ChunkID[:]); err != nil {
		file.Close()
		os.Remove(filePath) // Clean up on error
		return nil, fmt.Errorf("failed to write ChunkID: %w", err)
	}

	// Write chunk size (will be updated later)
	if err := binary.Write(file, binary.LittleEndian, header.ChunkSize); err != nil {
		file.Close()
		os.Remove(filePath) // Clean up on error
		return nil, fmt.Errorf("failed to write ChunkSize: %w", err)
	}

	// Write format
	if _, err := file.Write(header.Format[:]); err != nil {
		file.Close()
		os.Remove(filePath) // Clean up on error
		return nil, fmt.Errorf("failed to write Format: %w", err)
	}

	// Write fmt subchunk
	if _, err := file.Write(header.Subchunk1ID[:]); err != nil {
		file.Close()
		os.Remove(filePath) // Clean up on error
		return nil, fmt.Errorf("failed to write Subchunk1ID: %w", err)
	}

	if err := binary.Write(file, binary.LittleEndian, header.Subchunk1Size); err != nil {
		file.Close()
		os.Remove(filePath) // Clean up on error
		return nil, fmt.Errorf("failed to write Subchunk1Size: %w", err)
	}

	if err := binary.Write(file, binary.LittleEndian, header.AudioFormat); err != nil {
		file.Close()
		os.Remove(filePath) // Clean up on error
		return nil, fmt.Errorf("failed to write AudioFormat: %w", err)
	}

	if err := binary.Write(file, binary.LittleEndian, header.NumChannels); err != nil {
		file.Close()
		os.Remove(filePath) // Clean up on error
		return nil, fmt.Errorf("failed to write NumChannels: %w", err)
	}

	if err := binary.Write(file, binary.LittleEndian, header.SampleRate); err != nil {
		file.Close()
		os.Remove(filePath) // Clean up on error
		return nil, fmt.Errorf("failed to write SampleRate: %w", err)
	}

	if err := binary.Write(file, binary.LittleEndian, header.ByteRate); err != nil {
		file.Close()
		os.Remove(filePath) // Clean up on error
		return nil, fmt.Errorf("failed to write ByteRate: %w", err)
	}

	if err := binary.Write(file, binary.LittleEndian, header.BlockAlign); err != nil {
		file.Close()
		os.Remove(filePath) // Clean up on error
		return nil, fmt.Errorf("failed to write BlockAlign: %w", err)
	}

	if err := binary.Write(file, binary.LittleEndian, header.BitsPerSample); err != nil {
		file.Close()
		os.Remove(filePath) // Clean up on error
		return nil, fmt.Errorf("failed to write BitsPerSample: %w", err)
	}

	// Write data subchunk
	if _, err := file.Write(header.Subchunk2ID[:]); err != nil {
		file.Close()
		os.Remove(filePath) // Clean up on error
		return nil, fmt.Errorf("failed to write Subchunk2ID: %w", err)
	}

	// Write Subchunk2Size (will be updated later)
	if err := binary.Write(file, binary.LittleEndian, header.Subchunk2Size); err != nil {
		file.Close()
		os.Remove(filePath) // Clean up on error
		return nil, fmt.Errorf("failed to write Subchunk2Size: %w", err)
	}

	// Create the stream
	stream := &WhisperStream{
		filePath:    filePath,
		file:        file, // Store the file handle
		results:     make(chan Result, 10),
		ctx:         w.ctx,
		transcriber: w,
		language:    language,   // Store per-stream language
		transcribe:  transcribe, // Store transcribe flag
	}

	log.Printf("Whisper stream created: %s (language: %s, transcribe: %v)", fileName, language, transcribe)
	return stream, nil
}

// Results returns a channel that will receive the transcription results
func (ws *WhisperStream) Results() <-chan Result {
	return ws.results
}

// Close processes the audio file with Whisper and sends the result
func (ws *WhisperStream) Close() error {
	ws.mu.Lock()
	if ws.isClosed {
		ws.mu.Unlock()
		return nil
	}
	ws.isClosed = true
	ws.mu.Unlock()

	// Flush any buffered data to disk
	if err := ws.file.Sync(); err != nil {
		log.Printf("Warning: failed to sync file: %v", err)
	}

	// Get current file size
	fileInfo, err := ws.file.Stat()
	if err != nil {
		ws.file.Close()
		os.Remove(ws.filePath) // Clean up on error
		return fmt.Errorf("failed to get file info: %w", err)
	}

	// Calculate sizes
	fileSize := uint32(fileInfo.Size())

	// Check if we have enough data for a valid WAV file
	if fileSize < 44 {
		ws.file.Close()
		os.Remove(ws.filePath) // Clean up incomplete file
		return fmt.Errorf("file too small for WAV header: %d bytes", fileSize)
	}

	audioDataSize := fileSize - 44 // 44 bytes for WAV header

	// Update chunk size (file size - 8) at position 4
	chunkSize := fileSize - 8

	// Seek to position 4 (after ChunkID)
	if _, err := ws.file.Seek(4, 0); err != nil {
		ws.file.Close()
		os.Remove(ws.filePath) // Clean up on error
		return fmt.Errorf("failed to seek to ChunkSize position: %w", err)
	}

	if err := binary.Write(ws.file, binary.LittleEndian, chunkSize); err != nil {
		ws.file.Close()
		os.Remove(ws.filePath) // Clean up on error
		return fmt.Errorf("failed to update chunk size: %w", err)
	}

	// Seek to Subchunk2Size position (40 bytes from start)
	if _, err := ws.file.Seek(40, 0); err != nil {
		ws.file.Close()
		os.Remove(ws.filePath) // Clean up on error
		return fmt.Errorf("failed to seek to Subchunk2Size: %w", err)
	}

	// Update Subchunk2Size (audio data size)
	if err := binary.Write(ws.file, binary.LittleEndian, audioDataSize); err != nil {
		ws.file.Close()
		os.Remove(ws.filePath) // Clean up on error
		return fmt.Errorf("failed to update Subchunk2Size: %w", err)
	}

	// Flush the header updates to disk
	if err := ws.file.Sync(); err != nil {
		log.Printf("Warning: failed to sync header updates: %v", err)
	}

	// Close file
	if err := ws.file.Close(); err != nil {
		os.Remove(ws.filePath) // Clean up on error
		return fmt.Errorf("failed to close file: %w", err)
	}

	// Check if audio file has content
	if fileSize == 44 {
		log.Printf("Warning: Audio file is empty (only header), skipping transcription")
		// Clean up empty file
		os.Remove(ws.filePath)
		close(ws.results)
		return nil
	}

	// Check if transcription is enabled
	if !ws.transcribe {
		// Record only mode - just return the audio file info
		log.Printf("Record only mode - skipping transcription for: %s", ws.filePath)
		ws.results <- Result{
			Text:       "Recording saved (transcription disabled)",
			Confidence: 1.0,
			Final:      true,
			AudioFile:  ws.filePath,
		}
		close(ws.results)
		log.Printf("Recording completed: %s (Size: %d bytes, Audio: %d bytes)", filepath.Base(ws.filePath), fileSize, audioDataSize)
		return nil
	}

	// Transcribe audio using Whisper
	text, textFile, err := ws.transcribeAudio(ws.filePath)
	if err != nil {
		log.Printf("Error transcribing audio: %v", err)
		// Send error result but don't fail the stream
		ws.results <- Result{
			Text:       fmt.Sprintf("Transcription error: %v", err),
			Confidence: 0.0,
			Final:      true,
			AudioFile:  ws.filePath,
		}
	} else {
		// Send successful transcription result
		ws.results <- Result{
			Text:       text,
			Confidence: 0.9, // Whisper doesn't provide confidence scores
			Final:      true,
			AudioFile:  ws.filePath,
			TextFile:   textFile,
		}
	}

	// Clean up temporary file based on retention flags
	if !ws.transcriber.keepWav {
		if err := os.Remove(ws.filePath); err != nil {
			log.Printf("Warning: Failed to remove temporary file %s: %v", ws.filePath, err)
		}
	} else {
		log.Printf("Keeping WAV file: %s", ws.filePath)
	}

	close(ws.results)
	log.Printf("Whisper transcription completed: %s (Size: %d bytes, Audio: %d bytes)", filepath.Base(ws.filePath), fileSize, audioDataSize)
	return nil
}

// Write writes audio data to a temporary WAV file
func (ws *WhisperStream) Write(buffer []byte) (int, error) {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	if ws.isClosed {
		return 0, fmt.Errorf("stream is closed")
	}

	// Log audio data received
	//log.Printf("Received %d bytes of audio data for file: %s", len(buffer), filepath.Base(ws.filePath))

	// Write audio data directly to the stored file handle
	written, err := ws.file.Write(buffer)
	if err != nil {
		return written, fmt.Errorf("failed to write audio data: %w", err)
	}

	// Ensure data is written to disk
	if err := ws.file.Sync(); err != nil {
		log.Printf("Warning: failed to sync audio data: %v", err)
	}

	//log.Printf("Wrote %d bytes to audio file: %s", written, filepath.Base(ws.filePath))
	return written, nil
}

// transcribeAudio runs Whisper on the audio file and returns the transcription
func (ws *WhisperStream) transcribeAudio(audioPath string) (string, string, error) {
	// Check if Whisper is available
	if ws.transcriber.whisperPath == "" {
		return "", "", fmt.Errorf("whisper executable not found, please install whisper-ctranslate2 or set WHISPER_PATH")
	}

	// Use stream's language (which may override transcriber's default)
	language := ws.language
	if language == "" {
		language = ws.transcriber.language
	}

	log.Printf("Transcribing audio file: %s to output directory: %s (language: %s)", audioPath, ws.transcriber.tempDir, language)
	// Prepare Whisper command
	args := []string{
		"--model", ws.transcriber.modelPath,
		"--output_dir", ws.transcriber.tempDir,
		"--output_format", "txt",
		"--task", "transcribe",
		"--temperature", "0.0", // Deterministic output
	}

	// Add language parameter if specified (not "auto")
	if language != "" && language != "auto" {
		args = append(args, "--language", language)
	}

	// Add the audio file path
	args = append(args, audioPath)

	// Execute Whisper
	cmd := exec.CommandContext(ws.ctx, ws.transcriber.whisperPath, args...)
	// cmd.Dir = ws.transcriber.tempDir // Do not change dir, as audioPath is relative to project root

	// Capture output
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", "", fmt.Errorf("whisper execution failed: %w, output: %s", err, string(output))
	}

	// Read the transcription result
	outputFile := audioPath[:len(audioPath)-4] + ".txt" // Replace .wav with .txt
	content, err := os.ReadFile(outputFile)
	if err != nil {
		// Log the command output if reading the file fails, to help debug why it wasn't created
		log.Printf("Whisper command output: %s", string(output))
		return "", "", fmt.Errorf("failed to read transcription output: %w", err)
	}

	// Clean up output file based on retention flags
	if !ws.transcriber.keepTxt {
		if err := os.Remove(outputFile); err != nil {
			log.Printf("Warning: Failed to remove output file %s: %v", outputFile, err)
		}
	} else {
		log.Printf("Keeping TXT file: %s", outputFile)
	}

	// Return transcription text
	text := string(content)
	if text == "" {
		return "", outputFile, fmt.Errorf("transcription result is empty")
	}

	return text, outputFile, nil
}

// findWhisperExecutable searches for Whisper executable using "which" command first
func findWhisperExecutable() string {
	// Common Whisper executable names (in priority order)
	executables := []string{
		"whisper-ctranslate2", // Highest priority - fastest implementation
		"whisper",             // Original OpenAI Whisper
		"whisper.cpp",         // C++ implementation
		"whisper-ai",          // Alternative implementation
	}

	// Try "which" command first (most reliable)
	for _, execName := range executables {
		// Use "which" command to find executable in PATH
		cmd := exec.Command("which", execName)
		if output, err := cmd.Output(); err == nil {
			path := string(output)
			// Remove trailing newline
			path = strings.TrimSpace(path)
			if path != "" {
				log.Printf("Found Whisper executable using 'which': %s", path)
				return path
			}
		}
	}

	// Fallback: Check PATH using exec.LookPath
	for _, execName := range executables {
		if path, err := exec.LookPath(execName); err == nil {
			log.Printf("Found Whisper executable using exec.LookPath: %s", path)
			return path
		}
	}

	// Last resort: Check common installation paths
	commonPaths := []string{
		"/usr/local/bin",
		"/usr/bin",
		"/opt/homebrew/bin", // macOS Homebrew
		"/usr/local/opt/whisper/bin",
		"./bin",
		"./whisper",
	}

	for _, path := range commonPaths {
		for _, execName := range executables {
			fullPath := filepath.Join(path, execName)
			if _, err := os.Stat(fullPath); err == nil {
				log.Printf("Found Whisper executable in common path: %s", fullPath)
				return fullPath
			}
		}
	}

	log.Printf("No Whisper executable found in PATH or common locations")
	return ""
}

// findWhisperModel searches for Whisper models in common locations
func findWhisperModel() string {
	// Common model paths - prioritize whisper-ctranslate2 default location
	modelPaths := []string{
		"~/.cache/whisper", // whisper-ctranslate2 default location
		"./models",
		"./whisper-models",
		"/usr/local/share/whisper",
		"/opt/whisper/models",
	}

	// Common model names (from smallest to largest)
	models := []string{
		"tiny.en",
		"tiny",
		"base.en",
		"base",
		"small.en",
		"small",
		"medium.en",
		"medium",
		"large-v2",
		"large-v3",
	}

	// Check each path for models
	for _, modelPath := range modelPaths {
		// Expand home directory
		if modelPath[:2] == "~/" {
			home, err := os.UserHomeDir()
			if err == nil {
				modelPath = filepath.Join(home, modelPath[2:])
			}
		}

		for _, model := range models {
			fullPath := filepath.Join(modelPath, model)
			if _, err := os.Stat(fullPath); err == nil {
				log.Printf("Found Whisper model: %s", fullPath)
				return fullPath
			}
		}
	}

	log.Printf("No Whisper model found in common locations")
	return ""
}

// NewWhisperTranscriber creates a new instance of the transcribe.Service that uses Whisper
func NewWhisperTranscriber(ctx context.Context, modelPath, whisperPath, tempDir, language string, keepWav, keepTxt bool) (Service, error) {
	// Use provided paths or try to find them automatically
	if whisperPath == "" {
		whisperPath = findWhisperExecutable()
		if whisperPath == "" {
			return nil, fmt.Errorf("whisper executable not found, please install whisper-ctranslate2 or set WHISPER_PATH")
		}
	}

	if modelPath == "" {
		modelPath = findWhisperModel()
		if modelPath == "" {
			modelPath = "small"
		}
	}

	if tempDir == "" {
		tempDir = "./output"
	}

	// Default language to auto-detect if not specified
	if language == "" {
		language = "auto"
	}

	// Create temp directory if it doesn't exist
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %w", err)
	}

	// Verify Whisper executable
	if _, err := os.Stat(whisperPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("whisper executable not found at: %s", whisperPath)
	}

	log.Printf("Whisper transcriber initialized with model: %s, executable: %s, language: %s", modelPath, whisperPath, language)

	return &WhisperTranscriber{
		modelPath:   modelPath,
		whisperPath: whisperPath,
		tempDir:     tempDir,
		language:    language,
		ctx:         ctx,
		keepWav:     keepWav,
		keepTxt:     keepTxt,
	}, nil
}
