# Whisper Speech Recognition Setup

This document explains how to set up and use the Whisper transcriber service with the WebRTC transcriber application.

## Overview

The `WhisperTranscriber` implements the `transcribe.Service` interface using OpenAI's Whisper model for local speech recognition. Whisper provides high-quality speech recognition without sending audio data to external services, making it ideal for privacy-conscious applications and offline use.

## Prerequisites

1. **Whisper Executable**: A Whisper implementation (e.g., whisper-ctranslate2, whisper.cpp)
2. **Whisper Model**: A pre-trained Whisper model file
3. **System Resources**: Sufficient CPU/GPU resources for real-time transcription

## Whisper Implementations

### 1. whisper-ctranslate2 (Recommended)
Fastest and most efficient implementation:

```bash
# Install via pip
pip install whisper-ctranslate2

# Or via conda
conda install -c conda-forge whisper-ctranslate2
```

### 2. whisper.cpp
C++ implementation with good performance:

```bash
# Clone and build
git clone https://github.com/ggerganov/whisper.cpp.git
cd whisper.cpp
make

# The executable will be in the current directory
```

### 3. Original OpenAI Whisper
Python implementation (slower but more features):

```bash
pip install openai-whisper
```

## Model Download

Whisper models come in different sizes. Choose based on your accuracy vs. speed requirements:

### Model Sizes (from smallest to largest)
- **tiny.en/tiny**: ~39MB, fastest, English-only or multilingual
- **base.en/base**: ~74MB, good balance of speed and accuracy
- **small.en/small**: ~244MB, better accuracy, reasonable speed
- **medium.en/medium**: ~769MB, high accuracy, slower
- **large-v2/large-v3**: ~1550MB, best accuracy, slowest

### Download Models

```bash
# Using whisper-ctranslate2 (Recommended)
# Models are downloaded automatically on first use
whisper-ctranslate2 --model tiny --output_dir ./output /path/to/any/audio.wav
whisper-ctranslate2 --model base --output_dir ./output /path/to/any/audio.wav
whisper-ctranslate2 --model small --output_dir ./output /path/to/any/audio.wav

# Using original Whisper
whisper --model tiny --download_root ./models
whisper --model base --download_root ./models
whisper --model small --download_root ./models

# Manual download (models are stored in ~/.cache/whisper by default)
# Copy models from cache to desired location
cp -r ~/.cache/whisper/* ./models/
```

## Quick Setup Guide

### Step 1: Install Whisper
```bash
# Install whisper-ctranslate2 (recommended)
pip install whisper-ctranslate2

# Or install original Whisper
pip install openai-whisper
```

### Step 2: Download a Model
```bash
# Create a sample audio file or use any existing audio file
# Then run whisper to download the model automatically
whisper-ctranslate2 --model tiny --output_dir ./output /path/to/any/audio.wav
```

### Step 3: Set Environment Variables
```bash
# Set the model path (use the cache location where models are stored)
export WHISPER_MODEL_PATH="$HOME/.cache/whisper/tiny"

# Optionally set the whisper executable path
export WHISPER_PATH="/usr/local/bin/whisper-ctranslate2"
```

### Step 4: Test the Setup
```bash
# Run your WebRTC transcriber
./webrtc-transcriber
```

## Configuration

### Environment Variables

Set the following environment variables:

```bash
# Whisper model path (required)
export WHISPER_MODEL_PATH="/path/to/whisper/model"

# Whisper executable path (optional, auto-detected if not set)
export WHISPER_PATH="/path/to/whisper/executable"

# Examples:
export WHISPER_MODEL_PATH="./models/tiny"
export WHISPER_PATH="/usr/local/bin/whisper-ctranslate2"
```

### Running the Application

The Whisper service is automatically selected when the environment variables are set:

```bash
# Whisper will be used automatically
./webrtc-transcriber

# Or with custom port
./webrtc-transcriber --http.port=8080
```

## Service Selection Priority

The application automatically selects services in this order:
1. **Google Speech** (if `--google.cred` flag is provided)
2. **Azure Speech** (if `AZURE_SPEECH_KEY` and `AZURE_SPEECH_REGION` are set)
3. **Baidu Speech** (if `BAIDU_*` environment variables are set)
4. **Xunfei** (if `XUNFEI_*` environment variables are set)
5. **Whisper** (if `WHISPER_MODEL_PATH` or `WHISPER_PATH` are set) ⭐ **NEW!**
6. **Recorder** (fallback - no credentials needed)

## Features

### Speech Recognition Capabilities
- **High Accuracy**: State-of-the-art speech recognition models
- **Multiple Languages**: 100+ languages supported
- **Offline Processing**: No internet connection required
- **Privacy-First**: Audio data never leaves your system
- **Customizable**: Different model sizes for speed/accuracy trade-offs

### Audio Format Support
- **Sample Rate**: 16kHz recommended (auto-resampled if needed)
- **Channels**: Mono and stereo
- **Codecs**: WAV, MP3, OGG, FLAC, M4A
- **Bit Depth**: 16-bit recommended

### Model Configuration
- **Language Detection**: Automatic language detection
- **Task Types**: Transcription and translation
- **Temperature Control**: Adjustable randomness in output
- **Output Formats**: Text, JSON, SRT, VTT

## Performance Considerations

### Hardware Requirements

#### CPU-Only Processing
- **tiny/base models**: Good for real-time on modern CPUs
- **small/medium models**: May have latency on slower CPUs
- **large models**: Not recommended for real-time without GPU

#### GPU Acceleration
- **CUDA**: NVIDIA GPUs with CUDA support
- **OpenCL**: AMD and Intel GPUs
- **Metal**: Apple Silicon and AMD GPUs on macOS

### Optimization Tips
1. **Model Selection**: Use smaller models for real-time applications
2. **Batch Processing**: Process multiple audio chunks together
3. **GPU Memory**: Ensure sufficient GPU memory for larger models
4. **CPU Cores**: More CPU cores improve parallel processing

## Installation Examples

### macOS (Homebrew)
```bash
# Install whisper-ctranslate2
brew install whisper-ctranslate2

# Download a model (models are downloaded automatically on first use)
whisper-ctranslate2 --model tiny --output_dir ~/output /path/to/any/audio.wav

# Set environment variables
export WHISPER_MODEL_PATH="$HOME/.cache/whisper/tiny"
export WHISPER_PATH="/opt/homebrew/bin/whisper-ctranslate2"
```

### Ubuntu/Debian
```bash
# Install dependencies
sudo apt update
sudo apt install python3-pip ffmpeg

# Install whisper-ctranslate2
pip3 install whisper-ctranslate2

# Download a model (models are downloaded automatically on first use)
whisper-ctranslate2 --model tiny --output_dir ~/output /path/to/any/audio.wav

# Set environment variables
export WHISPER_MODEL_PATH="$HOME/.cache/whisper/tiny"
export WHISPER_PATH="/usr/local/bin/whisper-ctranslate2"
```

### Windows
```bash
# Install via pip
pip install whisper-ctranslate2

# Download a model (models are downloaded automatically on first use)
whisper-ctranslate2 --model tiny --output_dir C:\output C:\path\to\any\audio.wav

# Set environment variables
set WHISPER_MODEL_PATH=%USERPROFILE%\.cache\whisper\tiny
set WHISPER_PATH=C:\Users\%USERNAME%\AppData\Local\Programs\Python\Python39\Scripts\whisper-ctranslate2.exe
```

## Troubleshooting

### Common Issues

1. **Whisper Executable Not Found**
   - Verify the executable path is correct
   - Check if Whisper is in your PATH
   - Use absolute paths if needed

2. **Model Not Found**
   - Verify the model path is correct
   - Check if the model directory contains the model files
   - Download the model if it doesn't exist

3. **Model Download Issues**
   - Models are downloaded automatically on first use with whisper-ctranslate2
   - Use `--model` flag with any audio file to trigger model download
   - Ensure you have sufficient disk space (models are 39MB to 1.5GB)
   - Check internet connection for model downloads
   - Models are stored in `~/.cache/whisper` by default

3. **Permission Denied**
   - Ensure the Whisper executable has execute permissions
   - Check if the model directory is readable
   - Verify temp directory permissions

4. **Out of Memory**
   - Use a smaller model (tiny, base, small)
   - Close other applications to free memory
   - Consider using CPU-only processing

5. **Slow Performance**
   - Use a smaller model for faster processing
   - Enable GPU acceleration if available
   - Optimize audio chunk sizes

### Debug Information

The service provides detailed logging:
```
2025/08/27 17:00:00 Whisper transcriber initialized with model: ./models/tiny, executable: /usr/local/bin/whisper-ctranslate2
2025/08/27 17:00:01 Whisper stream created: whisper_audio_1_20250827_170001.wav
2025/08/27 17:00:05 Whisper transcription completed: whisper_audio_1_20250827_170001.wav
```

### Performance Monitoring

Monitor system resources during transcription:
```bash
# CPU usage
top -p $(pgrep webrtc-transcriber)

# Memory usage
ps aux | grep webrtc-transcriber

# GPU usage (if using GPU acceleration)
nvidia-smi
```

## Security

### Best Practices
1. **Local Processing**: Audio data never leaves your system
2. **Model Verification**: Download models from official sources
3. **File Permissions**: Restrict access to model files
4. **Network Isolation**: No external network calls required

### Privacy Benefits
- **Data Sovereignty**: Complete control over your audio data
- **No Cloud Dependencies**: Works completely offline
- **Compliance**: Meets strict privacy requirements
- **Audit Trail**: Full control over data processing

## Integration Examples

### With Docker
```dockerfile
FROM python:3.9-slim

# Install system dependencies
RUN apt-get update && apt-get install -y ffmpeg

# Install Whisper
RUN pip install whisper-ctranslate2

# Download model (models are downloaded automatically on first use)
RUN whisper-ctranslate2 --model tiny --output_dir /opt/output /tmp/sample.wav

# Set environment variables
ENV WHISPER_MODEL_PATH=/opt/whisper-models/tiny
ENV WHISPER_PATH=/usr/local/bin/whisper-ctranslate2

# Copy application
COPY webrtc-transcriber /usr/local/bin/
```

### With Kubernetes
```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: whisper-config
data:
  WHISPER_MODEL_PATH: "/opt/whisper-models/tiny"
  WHISPER_PATH: "/usr/local/bin/whisper-ctranslate2"
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
        envFrom:
        - configMapRef:
            name: whisper-config
        volumeMounts:
        - name: whisper-models
          mountPath: /opt/whisper-models
      volumes:
      - name: whisper-models
        persistentVolumeClaim:
          claimName: whisper-models-pvc
```

## Comparison with Other Services

### vs. Cloud Services (Google, Azure, Baidu)
- **Privacy**: ✅ No data sent to external services
- **Cost**: ✅ Free to use (just computational resources)
- **Latency**: ⚠️ Higher latency due to local processing
- **Accuracy**: ✅ Comparable or better accuracy
- **Offline**: ✅ Works completely offline

### vs. Recorder
- **Transcription**: ✅ Actual transcription vs. just recording
- **Privacy**: ✅ Local processing vs. local storage
- **Resources**: ⚠️ Higher computational requirements
- **Setup**: ⚠️ More complex setup required

## Support

For issues with Whisper:
- **whisper-ctranslate2**: [GitHub Repository](https://github.com/guillaumekln/faster-whisper)
- **whisper.cpp**: [GitHub Repository](https://github.com/ggerganov/whisper.cpp)
- **Original Whisper**: [OpenAI Documentation](https://github.com/openai/whisper)
- **Community**: [Whisper Community Discussions](https://github.com/openai/whisper/discussions)

## Migration from Other Services

### From Cloud Services
- Install Whisper and download models
- Set `WHISPER_MODEL_PATH` and optionally `WHISPER_PATH`
- Remove cloud service environment variables
- No code changes required

### From Recorder
- Install Whisper and download models
- Set Whisper environment variables
- Remove `RECORDER_OUTPUT_DIR` if not needed
- No code changes required

The Whisper service maintains the same interface, so no code changes are required in your application.

## Best Practices for Whisper

### Model Selection
- **Development/Testing**: Use `tiny` or `base` models
- **Production (Speed)**: Use `tiny` or `base` models
- **Production (Accuracy)**: Use `small` or `medium` models
- **Research/High Accuracy**: Use `large` models

### Audio Optimization
- **Sample Rate**: Use 16kHz for optimal performance
- **Chunk Size**: Balance between latency and efficiency
- **Format**: WAV format provides best compatibility
- **Quality**: Higher quality audio improves accuracy

### System Optimization
- **GPU Acceleration**: Enable if available for better performance
- **Memory Management**: Monitor memory usage with larger models
- **CPU Cores**: More cores improve parallel processing
- **Storage**: Ensure sufficient space for models and temporary files
