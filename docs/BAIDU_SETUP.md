# Baidu Speech Recognition Setup

This document explains how to set up and use the Baidu Speech Recognition transcriber service with the WebRTC transcriber application.

## Overview

The `BaiduTranscriber` implements the `transcribe.Service` interface using Baidu Speech Recognition API (百度语音识别) for high-quality Chinese speech recognition. Baidu Speech provides excellent accuracy for Chinese languages and dialects, with competitive pricing for the Chinese market.

## Prerequisites

1. **Baidu AI Platform Account**: You need a Baidu AI Platform account
2. **Speech Recognition Service**: Speech recognition service must be enabled
3. **API Credentials**: App ID, API Key, and Secret Key from your Baidu AI Platform

## Setup Steps

### 1. Create Baidu AI Platform Account

1. **Sign up**: Go to [ai.baidu.com](https://ai.baidu.com)
2. **Create Application**: Click "Create Application" in the console
3. **Select Services**: Choose "Speech Recognition" (语音识别)
4. **Get Credentials**: Note down your App ID, API Key, and Secret Key

### 2. Enable Speech Recognition

1. **Navigate to Console**: Go to your application dashboard
2. **Enable Service**: Make sure Speech Recognition is enabled
3. **Check Quota**: Verify your service quota and billing status

## Configuration

### Environment Variables

Set the following environment variables:

```bash
# Baidu Speech Recognition credentials
export BAIDU_APP_ID="your_app_id_here"
export BAIDU_API_KEY="your_api_key_here"
export BAIDU_SECRET_KEY="your_secret_key_here"

# Example:
export BAIDU_APP_ID="12345678"
export BAIDU_API_KEY="a1b2c3d4e5f6g7h8i9j0"
export BAIDU_SECRET_KEY="k1l2m3n4o5p6q7r8s9t0"
```

### Running the Application

The Baidu Speech service is automatically selected when the environment variables are set:

```bash
# Baidu Speech will be used automatically
./webrtc-transcriber

# Or with custom port
./webrtc-transcriber --http.port=8080
```

## Service Selection Priority

The application automatically selects services in this order:
1. **Google Speech** (if `--google.cred` flag is provided)
2. **Azure Speech** (if `AZURE_SPEECH_KEY` and `AZURE_SPEECH_REGION` are set)
3. **Baidu Speech** (if `BAIDU_*` environment variables are set) ⭐ **NEW!**
4. **Xunfei** (if `XUNFEI_*` environment variables are set)
5. **Recorder** (fallback - no credentials needed)

## Features

### Speech Recognition Capabilities
- **Chinese Language Support**: Excellent support for Mandarin Chinese
- **Multiple Dialects**: Support for various Chinese regional dialects
- **Real-time Streaming**: Low-latency transcription
- **High Accuracy**: State-of-the-art Chinese speech recognition
- **Cost-effective**: Competitive pricing for Chinese markets

### Audio Format Support
- **Sample Rate**: 8kHz, 16kHz (recommended)
- **Channels**: Mono (single channel)
- **Codecs**: PCM, WAV
- **Bit Depth**: 16-bit

### Language Models
- **Mandarin Chinese**: Standard Mandarin (dev_pid: 1537)
- **Cantonese**: Guangdong dialect support
- **Other Dialects**: Various regional Chinese dialects

## Pricing

Baidu Speech Recognition offers competitive pricing:

### Free Tier
- **Daily Quota**: 500 requests per day
- **Features**: Standard speech recognition
- **Best for**: Development, testing, low-volume usage

### Paid Tiers
- **Pay-per-use**: Very competitive pricing
- **Volume Discounts**: Available for high-volume usage
- **Enterprise Plans**: Custom pricing for large deployments

## Troubleshooting

### Common Issues

1. **Authentication Failed**
   - Verify your API key and secret key are correct
   - Ensure the credentials haven't expired
   - Check if the Speech Recognition service is enabled

2. **Quota Exceeded**
   - Check your daily quota in Baidu AI Platform console
   - Consider upgrading your plan
   - Monitor usage statistics

3. **Connection Issues**
   - Verify internet connectivity
   - Check firewall settings
   - Ensure the Baidu API endpoints are accessible

4. **Audio Format Issues**
   - Verify audio is 16kHz mono PCM
   - Check audio file size and duration
   - Ensure proper audio encoding

### Debug Information

The service provides detailed logging:
```
2025/08/27 17:00:00 Using Baidu Speech service
2025/08/27 17:00:01 Baidu Speech API stream started
2025/08/27 17:00:05 Recognition result: 你好世界
```

### Error Codes

Common Baidu API error codes:
- **3300**: Authentication failed
- **3301**: Invalid API key
- **3302**: Service unavailable
- **3303**: Quota exceeded
- **3304**: Invalid audio format

## Performance Considerations

### Latency
- **Typical**: 200-500ms end-to-end latency
- **Factors**: Network latency, audio quality, server load
- **Optimization**: Use stable internet connection

### Throughput
- **Concurrent Streams**: Limited by your quota and plan
- **Audio Quality**: 16kHz mono PCM recommended for best performance
- **Network**: Stable internet connection required

## Security

### Best Practices
1. **Secure Storage**: Store API keys securely, never commit to version control
2. **Key Rotation**: Regularly rotate your API keys
3. **Access Control**: Use environment variables for configuration
4. **Network Security**: Use HTTPS for all API communications

### Compliance
- **Data Privacy**: Baidu AI Platform follows Chinese data protection regulations
- **Service Level**: Enterprise-grade reliability and uptime

## Integration Examples

### With Docker
```dockerfile
FROM golang:1.19-alpine
# ... other instructions
ENV BAIDU_APP_ID=your_app_id
ENV BAIDU_API_KEY=your_api_key
ENV BAIDU_SECRET_KEY=your_secret_key
```

### With Kubernetes
```yaml
apiVersion: v1
kind: Secret
metadata:
  name: baidu-speech-secret
type: Opaque
data:
  baidu-app-id: <base64-encoded-app-id>
  baidu-api-key: <base64-encoded-api-key>
  baidu-secret-key: <base64-encoded-secret-key>
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: webrtc-transcriber
spec:
  template:
    spec:
      containers:
      - name: transcriber
        env:
        - name: BAIDU_APP_ID
          valueFrom:
            secretKeyRef:
              name: baidu-speech-secret
              key: baidu-app-id
        - name: BAIDU_API_KEY
          valueFrom:
            secretKeyRef:
              name: baidu-speech-secret
              key: baidu-api-key
        - name: BAIDU_SECRET_KEY
          valueFrom:
            secretKeyRef:
              name: baidu-speech-secret
              key: baidu-secret-key
```

## Comparison with Other Services

### vs. Google Speech
- **Chinese Support**: Better Chinese language accuracy
- **Cost**: More cost-effective for Chinese markets
- **Global Coverage**: Limited to Chinese-speaking regions

### vs. Azure Speech
- **Chinese Focus**: Specialized for Chinese languages
- **Pricing**: Competitive pricing for Chinese markets
- **Features**: More limited feature set

### vs. Xunfei
- **Accuracy**: Comparable accuracy for Chinese
- **Pricing**: Similar pricing structure
- **Integration**: Easier integration with Baidu ecosystem

## Support

For issues with Baidu Speech Recognition:
- **Baidu AI Platform**: [ai.baidu.com](https://ai.baidu.com)
- **API Documentation**: [Speech Recognition API Docs](https://ai.baidu.com/ai-doc/SPEECH/Vk38lxily)
- **Developer Community**: [Baidu AI Developer Forum](https://ai.baidu.com/forum/)
- **Technical Support**: Available through Baidu AI Platform

## Migration from Other Services

### From Google Speech
- Set `BAIDU_*` environment variables
- Remove `--google.cred` flag
- No code changes required

### From Azure Speech
- Set Baidu environment variables
- Remove Azure environment variables
- No code changes required

### From Xunfei
- Set Baidu environment variables
- Remove Xunfei environment variables
- No code changes required

The Baidu Speech service maintains the same interface, so no code changes are required in your application.

## Best Practices for Chinese Speech Recognition

### Audio Quality
- **Sample Rate**: Use 16kHz for optimal accuracy
- **Noise Reduction**: Minimize background noise
- **Clear Speech**: Encourage clear pronunciation

### Language Settings
- **Dialect Selection**: Choose appropriate dialect for your region
- **Custom Vocabulary**: Add domain-specific terms if needed
- **Context Awareness**: Provide context when possible

### Performance Optimization
- **Chunk Size**: Use appropriate audio chunk sizes
- **Streaming**: Leverage real-time streaming for better user experience
- **Error Handling**: Implement robust error handling for network issues
