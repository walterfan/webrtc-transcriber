# Recorder Transcriber Setup

This document explains how to set up and use the Recorder transcriber service, which records audio tracks to local WAV files instead of performing speech recognition.

## Overview

The `RecorderTranscriber` implements the `transcribe.Service` interface and records incoming audio streams to local WAV files. This is useful for:

- **Audio Archiving**: Saving audio recordings for later analysis
- **Debugging**: Capturing audio to troubleshoot issues
- **Offline Processing**: Recording audio for batch processing later
- **Testing**: Verifying audio quality without transcription costs

## How It Works

1. **Audio Capture**: Receives audio data from WebRTC streams
2. **WAV File Creation**: Creates standard WAV files with proper headers
3. **File Naming**: Generates unique filenames with timestamps
4. **Result Return**: Returns the filename in the `Text` field of the `Result`

## Configuration

### Environment Variables

The recorder service can be configured using environment variables:

```bash
# Optional: Custom output directory (default: ./recordings)
export RECORDER_OUTPUT_DIR="/path/to/your/recordings"

# If not set, defaults to ./recordings in the current working directory
```

### Running the Application

The recorder service is automatically selected when no other transcription service credentials are provided:

```bash
# No credentials - will use Recorder service
./webrtc-transcriber

# With Google credentials - will use Google Speech
./webrtc-transcriber -google.cred=/path/to/google-credentials.json

# With Xunfei credentials - will use Xunfei service
export XUNFEI_APP_ID="your_app_id"
export XUNFEI_API_KEY="your_api_key"
export XUNFEI_API_SECRET="your_api_secret"
./webrtc-transcriber
```

## Features

### Audio Format
- **Encoding**: LINEAR16 (PCM)
- **Sample Rate**: 48kHz
- **Channels**: Mono (1 channel)
- **Bit Depth**: 16-bit
- **Format**: Standard WAV file

### File Management
- **Unique Naming**: `recording_YYYYMMDD_HHMMSS_NNN.wav`
- **Timestamp**: Includes date and time in filename
- **Counter**: Sequential numbering for multiple recordings
- **Directory Creation**: Automatically creates output directory if needed

### Error Handling
- **File Cleanup**: Removes incomplete files on errors
- **Resource Management**: Proper file handle cleanup
- **Concurrent Safety**: Thread-safe operations with mutex protection

## File Structure

### Output Directory
```
./recordings/
├── recording_20250827_161500_001.wav
├── recording_20250827_161520_002.wav
├── recording_20250827_161540_003.wav
└── ...
```

### WAV File Format
- **Header**: 44-byte WAV header with correct metadata
- **Audio Data**: Raw PCM audio samples
- **Sizes**: Automatically calculated and updated on close

## Usage Example

```go
import "github.com/walterfan/webrtc-transcriber/internal/transcribe"

ctx := context.Background()
recorder, err := transcribe.NewRecorderTranscriber(ctx, "./my_recordings")
if err != nil {
    log.Fatal(err)
}

stream, err := recorder.CreateStream()
if err != nil {
    log.Fatal(err)
}

// Write audio data
stream.Write(audioBuffer)

// Close stream to finalize file
stream.Close()

// Read results to get filename
for result := range stream.Results() {
    fmt.Printf("Recording saved as: %s\n", result.Text)
    fmt.Printf("Confidence: %.2f\n", result.Confidence)
    fmt.Printf("Final: %v\n", result.Final)
}
```

## Result Format

When using the recorder service, the `Result` struct contains:

```go
type Result struct {
    Text       string  `json:"text"`        // WAV filename (e.g., "recording_20250827_161500_001.wav")
    Confidence float32 `json:"confidence"`  // Always 1.0 (recording is always successful)
    Final      bool    `json:"final"`       // Always true (recording is complete)
}
```

## Performance Considerations

- **Disk Space**: Each recording consumes disk space proportional to audio length
- **I/O Performance**: File I/O operations may impact real-time performance
- **Concurrent Recording**: Multiple streams can record simultaneously
- **Memory Usage**: Minimal memory overhead, audio data written directly to disk

## Troubleshooting

### Common Issues

1. **Permission Denied**
   - Check write permissions for output directory
   - Ensure sufficient disk space
   - Verify directory path is accessible

2. **File Not Found**
   - Check if output directory was created
   - Verify environment variable settings
   - Check application working directory

3. **Corrupted WAV Files**
   - Ensure proper stream closure
   - Check for disk space issues
   - Verify audio format compatibility

### Debug Information

The service provides detailed logging:
```
2025/08/27 16:15:00 Started recording to: ./recordings/recording_20250827_161500_001.wav
2025/08/27 16:15:30 Recording completed: recording_20250827_161500_001.wav (Size: 1440000 bytes)
```

## Integration with WebRTC

The recorder service integrates seamlessly with the existing WebRTC infrastructure:

1. **Audio Stream**: Receives Opus-encoded audio from WebRTC
2. **Decoding**: Audio is decoded to PCM by the existing Opus decoder
3. **Recording**: PCM data is written to WAV files
4. **Results**: Filename is sent back through the existing DataChannel

## Use Cases

### Development and Testing
- Capture audio samples for analysis
- Debug audio quality issues
- Test different audio configurations

### Production Recording
- Archive important conversations
- Compliance and legal requirements
- Quality assurance and training

### Research and Analysis
- Audio pattern analysis
- Speech recognition training data
- Acoustic environment studies

## Limitations

- **No Transcription**: Audio is recorded but not converted to text
- **File Size**: WAV files can be large for long recordings
- **Disk Space**: Requires sufficient storage for recordings
- **No Compression**: Uses uncompressed WAV format for compatibility

## Future Enhancements

Potential improvements for the recorder service:
- **Audio Compression**: Support for MP3, OGG, or FLAC formats
- **Streaming**: Real-time audio streaming to external services
- **Metadata**: Additional file metadata (duration, bitrate, etc.)
- **Cleanup**: Automatic file rotation and cleanup policies
- **Quality Settings**: Configurable audio quality parameters

## Support

For issues with the recorder service:
- Check application logs for error messages
- Verify file system permissions and disk space
- Ensure proper stream closure in your application
- Review the WAV file format specification if needed
