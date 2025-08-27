# Azure Speech Service Setup

This document explains how to set up and use the Azure Speech transcriber service with the WebRTC transcriber application.

## Overview

The `AzureTranscriber` implements the `transcribe.Service` interface using Microsoft Azure Speech Service for high-quality speech recognition. Azure Speech Service provides excellent accuracy, multiple language support, and enterprise-grade reliability.

## Prerequisites

1. **Azure Account**: You need an Azure account (free tier available)
2. **Speech Service Resource**: A Speech service resource must be created in Azure Portal
3. **API Key and Region**: The subscription key and region from your Speech service resource

## Setup Steps

### 1. Create Azure Speech Service Resource

1. **Sign in to Azure Portal**: Go to [portal.azure.com](https://portal.azure.com)
2. **Create Resource**: Click "Create a resource" and search for "Speech service"
3. **Select Speech Service**: Choose "Speech service" from the results
4. **Configure Resource**:
   - **Subscription**: Select your Azure subscription
   - **Resource group**: Create new or use existing
   - **Region**: Choose a region close to your users (e.g., East US, West Europe)
   - **Name**: Give your resource a unique name
   - **Pricing tier**: Choose appropriate tier (F0 for free, S0 for standard)
5. **Review and Create**: Click "Review + create" then "Create"

### 2. Get Your Credentials

1. **Navigate to Resource**: Once created, go to your Speech service resource
2. **Get Keys**: In the left menu, click "Keys and Endpoint"
3. **Copy Key 1**: Copy the value of "Key 1"
4. **Note Region**: Copy the "Location/Region" value

## Configuration

### Environment Variables

Set the following environment variables:

```bash
# Azure Speech Service credentials
export AZURE_SPEECH_KEY="your_subscription_key_here"
export AZURE_SPEECH_REGION="your_region_here"

# Example:
export AZURE_SPEECH_KEY="a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6"
export AZURE_SPEECH_REGION="eastus"
```

### Running the Application

The Azure Speech service is automatically selected when the environment variables are set:

```bash
# Azure Speech will be used automatically
./webrtc-transcriber

# Or with custom port
./webrtc-transcriber --http.port=8080
```

## Service Selection Priority

The application automatically selects services in this order:
1. **Google Speech** (if `--google.cred` flag is provided)
2. **Azure Speech** (if `AZURE_SPEECH_KEY` and `AZURE_SPEECH_REGION` are set)
3. **Xunfei** (if `XUNFEI_*` environment variables are set)
4. **Recorder** (fallback - no credentials needed)

## Features

### Speech Recognition Capabilities
- **High Accuracy**: State-of-the-art speech recognition models
- **Multiple Languages**: Support for 100+ languages and dialects
- **Real-time Streaming**: Low-latency transcription
- **Custom Models**: Ability to train custom speech models
- **Speaker Identification**: Identify different speakers in conversations

### Audio Format Support
- **Sample Rate**: 8kHz, 16kHz, 32kHz, 48kHz
- **Channels**: Mono and stereo
- **Codecs**: PCM, WAV, MP3, OGG, FLAC
- **Bit Depth**: 8-bit, 16-bit, 24-bit

## Pricing

Azure Speech Service offers several pricing tiers:

### Free Tier (F0)
- **Monthly Hours**: 5 hours free per month
- **Features**: Standard speech recognition
- **Best for**: Development, testing, low-volume usage

### Standard Tier (S0)
- **Pay-per-use**: $16.00 per hour
- **Features**: All features including custom models
- **Best for**: Production applications

### Custom Speech
- **Training**: $6.00 per hour
- **Hosting**: $6.00 per hour
- **Features**: Custom speech models for domain-specific vocabulary

## Troubleshooting

### Common Issues

1. **Authentication Failed**
   - Verify your subscription key is correct
   - Ensure the key hasn't expired
   - Check if the Speech service resource is active

2. **Region Mismatch**
   - Verify the region matches your Speech service resource
   - Check for typos in the region name

3. **Quota Exceeded**
   - Check your Azure portal for usage statistics
   - Consider upgrading to a higher tier
   - Monitor usage with Azure Monitor

4. **Connection Issues**
   - Verify internet connectivity
   - Check firewall settings
   - Ensure the Speech service endpoint is accessible

### Debug Information

The service provides detailed logging:
```
2025/08/27 17:00:00 Using Azure Speech service (region: eastus)
2025/08/27 17:00:01 Azure Speech Service stream started
2025/08/27 17:00:05 Recognition result: Hello world
```

## Performance Considerations

### Latency
- **Typical**: 100-300ms end-to-end latency
- **Factors**: Network latency, audio quality, model complexity
- **Optimization**: Use regions close to your users

### Throughput
- **Concurrent Streams**: Limited by your pricing tier
- **Audio Quality**: Higher quality may improve accuracy but increase processing time
- **Network**: Stable internet connection required for optimal performance

## Security

### Best Practices
1. **Secure Storage**: Store API keys securely, never commit to version control
2. **Key Rotation**: Regularly rotate your subscription keys
3. **Access Control**: Use Azure RBAC to control access to Speech service
4. **Network Security**: Use Azure Private Link for secure connections

### Compliance
- **GDPR**: Azure Speech Service is GDPR compliant
- **HIPAA**: Available with appropriate Azure subscription
- **SOC**: SOC 1, 2, and 3 compliance available

## Integration Examples

### With Docker
```dockerfile
FROM golang:1.19-alpine
# ... other instructions
ENV AZURE_SPEECH_KEY=your_key
ENV AZURE_SPEECH_REGION=your_region
```

### With Kubernetes
```yaml
apiVersion: v1
kind: Secret
metadata:
  name: azure-speech-secret
type: Opaque
data:
  azure-speech-key: <base64-encoded-key>
  azure-speech-region: <base64-encoded-region>
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
        - name: AZURE_SPEECH_KEY
          valueFrom:
            secretKeyRef:
              name: azure-speech-secret
              key: azure-speech-key
        - name: AZURE_SPEECH_REGION
          valueFrom:
            secretKeyRef:
              name: azure-speech-secret
              key: azure-speech-region
```

## Support

For issues with Azure Speech Service:
- **Azure Documentation**: [Speech Service Documentation](https://docs.microsoft.com/en-us/azure/cognitive-services/speech-service/)
- **Azure Support**: Available through Azure portal
- **Community**: [Azure Speech Service Community](https://techcommunity.microsoft.com/t5/azure-cognitive-services/bd-p/AzureCognitiveServices)

## Migration from Other Services

### From Google Speech
- Set `AZURE_SPEECH_KEY` and `AZURE_SPEECH_REGION` environment variables
- Remove `--google.cred` flag
- No code changes required

### From Xunfei
- Set Azure environment variables
- Remove Xunfei environment variables
- No code changes required

The Azure Speech service maintains the same interface, so no code changes are required in your application.
