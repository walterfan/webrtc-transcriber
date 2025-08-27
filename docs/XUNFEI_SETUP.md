# Xunfei (IflyTek) Transcriber Setup

This document explains how to set up and use the Xunfei (讯飞) speech recognition service as an alternative to Google Speech-to-Text.

## Overview

The `IflyTekTranscriber` implements the `transcribe.Service` interface and uses Xunfei's WebSocket API for real-time speech recognition. It supports Chinese language recognition with features like dynamic correction and multi-dialect support.

## Prerequisites

1. **Xunfei Open Platform Account**: Sign up at [https://www.xfyun.cn/](https://www.xfyun.cn/)
2. **Speech Recognition Service**: Enable the "语音听写" (Speech Dictation) service
3. **API Credentials**: Get your AppID, API Key, and API Secret from the Xunfei console

## Configuration

### Environment Variables

Set the following environment variables with your Xunfei credentials:

```bash
export XUNFEI_APP_ID="your_app_id_here"
export XUNFEI_API_KEY="your_api_key_here"
export XUNFEI_API_SECRET="your_api_secret_here"
```

### Running the Application

The application will automatically use Xunfei if Google Speech credentials are not provided:

```bash
# Without Google credentials - will use Xunfei
./webrtc-transcriber

# With Google credentials - will use Google Speech
./webrtc-transcriber -google.cred=/path/to/google-credentials.json
```

## Features

### Language Support
- **Primary**: Chinese (zh_cn)
- **Dialects**: Supports 23+ Chinese dialects including:
  - Sichuan dialect (四川话)
  - Henan dialect (河南话)
  - Northeast dialect (东北话)
  - Cantonese (粤语)
  - Minnan dialect (闽南话)
  - Shandong dialect (山东话)
  - Guizhou dialect (贵州话)

### Audio Format
- **Encoding**: LINEAR16 (PCM)
- **Sample Rate**: 48kHz
- **Channels**: Mono (1 channel)
- **Format**: Raw audio data

### Advanced Features
- **Dynamic Correction**: Real-time result correction for better accuracy
- **Voice Activity Detection**: Automatic end-of-speech detection
- **Punctuation**: Automatic punctuation insertion
- **Real-time Streaming**: Low-latency transcription with partial results

## API Endpoints

The service uses Xunfei's WebSocket API:
- **Endpoint**: `wss://iat-api.xfyun.cn/v2/iat`
- **Protocol**: WebSocket over WSS (secure)
- **Authentication**: HMAC-SHA256 signature-based authentication

## Authentication Flow

1. **Timestamp Generation**: Current Unix timestamp
2. **Signature Creation**: HMAC-SHA256 of request line with API secret
3. **Authorization Header**: Base64-encoded authorization string
4. **URL Construction**: Query parameters for authentication

## Usage Example

```go
import "github.com/walterfan/webrtc-transcriber/internal/transcribe"

ctx := context.Background()
transcriber, err := transcribe.NewIflyTekTranscriber(
    ctx, 
    "your_app_id", 
    "your_api_key", 
    "your_api_secret"
)
if err != nil {
    log.Fatal(err)
}

stream, err := transcriber.CreateStream()
if err != nil {
    log.Fatal(err)
}

// Write audio data
stream.Write(audioBuffer)

// Read results
for result := range stream.Results() {
    fmt.Printf("Text: %s, Final: %v, Confidence: %.2f\n", 
        result.Text, result.Final, result.Confidence)
}

stream.Close()
```

## Error Handling

The service includes comprehensive error handling:
- **Connection Errors**: WebSocket connection failures
- **Authentication Errors**: Invalid credentials or signatures
- **API Errors**: Xunfei service errors with detailed messages
- **Audio Format Errors**: Unsupported audio format or parameters

## Performance Considerations

- **Latency**: Typically 100-300ms for real-time transcription
- **Concurrency**: Each stream maintains its own WebSocket connection
- **Memory**: Audio data is processed in chunks to minimize memory usage
- **Network**: Requires stable internet connection for WebSocket communication

## Troubleshooting

### Common Issues

1. **Authentication Failed**
   - Verify AppID, API Key, and API Secret
   - Check system clock synchronization
   - Ensure API service is enabled in Xunfei console

2. **Connection Timeout**
   - Check network connectivity
   - Verify firewall settings
   - Ensure WebSocket ports are accessible

3. **Audio Format Errors**
   - Verify audio is LINEAR16/PCM format
   - Check sample rate is 48kHz
   - Ensure mono channel configuration

### Debug Logging

Enable debug logging by setting the log level:
```bash
export LOG_LEVEL=debug
```

## Support

For Xunfei API support:
- **Documentation**: [https://www.xfyun.cn/doc/asr/voicedictation/API.html](https://www.xfyun.cn/doc/asr/voicedictation/API.html)
- **Community**: [https://www.xfyun.cn/community](https://www.xfyun.cn/community)
- **Console**: [https://console.xfyun.cn/](https://console.xfyun.cn/)

## Migration from Google Speech

To migrate from Google Speech to Xunfei:

1. **Update Environment**: Set Xunfei environment variables
2. **Remove Google Credentials**: Don't pass `-google.cred` flag
3. **Test Audio Format**: Ensure audio format compatibility
4. **Monitor Performance**: Compare latency and accuracy

The service maintains the same interface, so no code changes are required in your application.
