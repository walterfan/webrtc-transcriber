package transcribe

import (
	"context"
	"encoding/binary"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// RecorderTranscriber is the implementation of the transcribe.Service,
// it records audio tracks to local WAV files
type RecorderTranscriber struct {
	outputDir string
	ctx       context.Context
	mu        sync.Mutex
	counter   int
}

// RecorderStream implements the transcribe.Stream interface,
// it records audio data to a WAV file
type RecorderStream struct {
	file     *os.File
	results  chan Result
	ctx      context.Context
	fileName string
	filePath string
	mu       sync.Mutex
	isClosed bool
}

// WAV file header structure
type wavHeader struct {
	ChunkID       [4]byte // "RIFF"
	ChunkSize     uint32  // File size - 8
	Format        [4]byte // "WAVE"
	Subchunk1ID   [4]byte // "fmt "
	Subchunk1Size uint32  // 16 for PCM
	AudioFormat   uint16  // 1 for PCM
	NumChannels   uint16  // 1 for mono
	SampleRate    uint32  // 48000
	ByteRate      uint32  // SampleRate * NumChannels * BitsPerSample/8
	BlockAlign    uint16  // NumChannels * BitsPerSample/8
	BitsPerSample uint16  // 16
	Subchunk2ID   [4]byte // "data"
	Subchunk2Size uint32  // Size of audio data
}

// CreateStream creates a new recording stream
func (r *RecorderTranscriber) CreateStream() (Stream, error) {
	r.mu.Lock()
	r.counter++
	counter := r.counter
	r.mu.Unlock()

	// Generate unique filename with timestamp
	timestamp := time.Now().Format("20060102_150405")
	fileName := fmt.Sprintf("recording_%s_%03d.wav", timestamp, counter)
	filePath := filepath.Join(r.outputDir, fileName)

	// Create output directory if it doesn't exist
	if err := os.MkdirAll(r.outputDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create output directory: %w", err)
	}

	// Create WAV file
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

	stream := &RecorderStream{
		file:     file,
		results:  make(chan Result, 1), // Buffered channel to avoid blocking
		ctx:      r.ctx,
		fileName: fileName,
		filePath: filePath,
	}

	log.Printf("Started recording to: %s", filePath)
	return stream, nil
}

// Results returns a channel that will receive the recording result
func (rs *RecorderStream) Results() <-chan Result {
	return rs.results
}

// Close finalizes the WAV file and sends the result
func (rs *RecorderStream) Close() error {
	rs.mu.Lock()
	if rs.isClosed {
		rs.mu.Unlock()
		return nil
	}
	rs.isClosed = true
	rs.mu.Unlock()

	// Flush any buffered data to disk
	if err := rs.file.Sync(); err != nil {
		log.Printf("Warning: failed to sync file: %v", err)
	}

	// Get current file size
	fileInfo, err := rs.file.Stat()
	if err != nil {
		rs.file.Close()
		os.Remove(rs.filePath) // Clean up on error
		return fmt.Errorf("failed to get file info: %w", err)
	}

	// Calculate sizes
	fileSize := uint32(fileInfo.Size())

	// Check if we have enough data for a valid WAV file
	if fileSize < 44 {
		rs.file.Close()
		os.Remove(rs.filePath) // Clean up incomplete file
		return fmt.Errorf("file too small for WAV header: %d bytes", fileSize)
	}

	audioDataSize := fileSize - 44 // 44 bytes for WAV header

	// Update chunk size (file size - 8) at position 4
	chunkSize := fileSize - 8

	// Seek to position 4 (after ChunkID)
	if _, err := rs.file.Seek(4, 0); err != nil {
		rs.file.Close()
		os.Remove(rs.filePath) // Clean up on error
		return fmt.Errorf("failed to seek to ChunkSize position: %w", err)
	}

	if err := binary.Write(rs.file, binary.LittleEndian, chunkSize); err != nil {
		rs.file.Close()
		os.Remove(rs.filePath) // Clean up on error
		return fmt.Errorf("failed to update chunk size: %w", err)
	}

	// Seek to Subchunk2Size position (40 bytes from start)
	if _, err := rs.file.Seek(40, 0); err != nil {
		rs.file.Close()
		os.Remove(rs.filePath) // Clean up on error
		return fmt.Errorf("failed to seek to Subchunk2Size: %w", err)
	}

	// Update Subchunk2Size (audio data size)
	if err := binary.Write(rs.file, binary.LittleEndian, audioDataSize); err != nil {
		rs.file.Close()
		os.Remove(rs.filePath) // Clean up on error
		return fmt.Errorf("failed to update Subchunk2Size: %w", err)
	}

	// Flush the header updates to disk
	if err := rs.file.Sync(); err != nil {
		log.Printf("Warning: failed to sync header updates: %v", err)
	}

	// Close file
	if err := rs.file.Close(); err != nil {
		os.Remove(rs.filePath) // Clean up on error
		return fmt.Errorf("failed to close file: %w", err)
	}

	// Send result with filename
	rs.results <- Result{
		Text:       rs.fileName,
		Confidence: 1.0, // Recording is always successful
		Final:      true,
	}

	// Close results channel
	close(rs.results)

	log.Printf("Recording completed: %s (Size: %d bytes, Audio: %d bytes)", rs.fileName, fileSize, audioDataSize)

	// Validate the WAV file was created correctly
	if err := rs.validateWAVFile(); err != nil {
		log.Printf("Warning: WAV file validation failed: %v", err)
		// Don't return error here as the file was already closed
	}

	return nil
}

// validateWAVFile validates that the created WAV file has the correct structure
func (rs *RecorderStream) validateWAVFile() error {
	// Reopen file for validation
	file, err := os.Open(rs.filePath)
	if err != nil {
		return fmt.Errorf("failed to open file for validation: %w", err)
	}
	defer file.Close()

	// Read header manually to match how we wrote it
	var chunkID [4]byte
	if _, err := file.Read(chunkID[:]); err != nil {
		return fmt.Errorf("failed to read ChunkID: %w", err)
	}

	var chunkSize uint32
	if err := binary.Read(file, binary.LittleEndian, &chunkSize); err != nil {
		return fmt.Errorf("failed to read ChunkSize: %w", err)
	}

	var format [4]byte
	if _, err := file.Read(format[:]); err != nil {
		return fmt.Errorf("failed to read Format: %w", err)
	}

	// Skip to fmt subchunk
	if _, err := file.Seek(12, 0); err != nil {
		return fmt.Errorf("failed to seek to fmt subchunk: %w", err)
	}

	var subchunk1ID [4]byte
	if _, err := file.Read(subchunk1ID[:]); err != nil {
		return fmt.Errorf("failed to read Subchunk1ID: %w", err)
	}

	var subchunk1Size uint32
	if err := binary.Read(file, binary.LittleEndian, &subchunk1Size); err != nil {
		return fmt.Errorf("failed to read Subchunk1Size: %w", err)
	}

	var audioFormat uint16
	if err := binary.Read(file, binary.LittleEndian, &audioFormat); err != nil {
		return fmt.Errorf("failed to read AudioFormat: %w", err)
	}

	var numChannels uint16
	if err := binary.Read(file, binary.LittleEndian, &numChannels); err != nil {
		return fmt.Errorf("failed to read NumChannels: %w", err)
	}

	var sampleRate uint32
	if err := binary.Read(file, binary.LittleEndian, &sampleRate); err != nil {
		return fmt.Errorf("failed to read SampleRate: %w", err)
	}

	// Skip ByteRate and BlockAlign
	if _, err := file.Seek(32, 0); err != nil {
		return fmt.Errorf("failed to seek to BitsPerSample: %w", err)
	}

	var bitsPerSample uint16
	if err := binary.Read(file, binary.LittleEndian, &bitsPerSample); err != nil {
		return fmt.Errorf("failed to read BitsPerSample: %w", err)
	}

	// Skip to data subchunk
	if _, err := file.Seek(36, 0); err != nil {
		return fmt.Errorf("failed to seek to data subchunk: %w", err)
	}

	var subchunk2ID [4]byte
	if _, err := file.Read(subchunk2ID[:]); err != nil {
		return fmt.Errorf("failed to read Subchunk2ID: %w", err)
	}

	// Validate RIFF header
	if string(chunkID[:]) != "RIFF" {
		return fmt.Errorf("invalid RIFF header: %s", string(chunkID[:]))
	}

	// Validate WAVE format
	if string(format[:]) != "WAVE" {
		return fmt.Errorf("invalid WAVE format: %s", string(format[:]))
	}

	// Validate fmt subchunk
	if string(subchunk1ID[:]) != "fmt " {
		return fmt.Errorf("invalid fmt subchunk: %s", string(subchunk1ID[:]))
	}

	// Validate data subchunk
	if string(subchunk2ID[:]) != "data" {
		return fmt.Errorf("invalid data subchunk: %s", string(subchunk2ID[:]))
	}

	// Validate audio format (should be PCM = 1)
	if audioFormat != 1 {
		return fmt.Errorf("invalid audio format: %d (expected 1 for PCM)", audioFormat)
	}

	// Validate sample rate (should be 48000)
	if sampleRate != 48000 {
		return fmt.Errorf("invalid sample rate: %d (expected 48000)", sampleRate)
	}

	// Validate bits per sample (should be 16)
	if bitsPerSample != 16 {
		return fmt.Errorf("invalid bits per sample: %d (expected 16)", bitsPerSample)
	}

	// Validate channels (should be 1 for mono)
	if numChannels != 1 {
		return fmt.Errorf("invalid channel count: %d (expected 1)", numChannels)
	}

	log.Printf("WAV file validation passed for %s", rs.fileName)
	return nil
}

// Write writes audio data to the WAV file
func (rs *RecorderStream) Write(buffer []byte) (int, error) {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	if rs.isClosed {
		return 0, fmt.Errorf("stream is closed")
	}

	// Validate buffer size (should be even for 16-bit samples)
	if len(buffer)%2 != 0 {
		log.Printf("Warning: Odd buffer size %d, audio may be corrupted", len(buffer))
	}

	// Write audio data directly to file
	// Note: We assume the incoming audio is already in the correct format (16-bit PCM, 48kHz, mono)
	written, err := rs.file.Write(buffer)
	if err != nil {
		return written, fmt.Errorf("failed to write audio data: %w", err)
	}

	// Flush data to disk periodically to ensure it's written
	if written > 0 {
		if err := rs.file.Sync(); err != nil {
			log.Printf("Warning: failed to sync audio data: %v", err)
		}
	}

	return written, nil
}

// NewRecorderTranscriber creates a new instance of the transcribe.Service that records
// audio to local WAV files
func NewRecorderTranscriber(ctx context.Context, outputDir string) (Service, error) {
	if outputDir == "" {
		outputDir = "./recordings" // Default output directory
	}

	// Ensure output directory exists
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create output directory: %w", err)
	}

	return &RecorderTranscriber{
		outputDir: outputDir,
		ctx:       ctx,
		counter:   0,
	}, nil
}
