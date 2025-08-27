## WebRTC Recorder and Transcriber

This project provides a WebRTC-based audio recording and transcription service with support for multiple transcription backends.

### Features

- **Real-time Audio Streaming**: WebRTC-based audio capture and streaming
- **Multiple Transcription Services**: 
  - Google Speech-to-Text (cloud-based)
  - Azure Speech Service (enterprise-grade)
  - Baidu Speech Recognition (Chinese-optimized)
  - Xunfei (讯飞) Speech Recognition (optimized for Chinese)
  - Whisper Speech Recognition (local/offline)
  - Local Audio Recorder (WAV file output)
- **Flexible Configuration**: Automatic service selection based on available credentials
- **Cross-platform**: Works on Chrome, Firefox, and Safari

### Dependencies

The speech to text server only depends on [Go 1.12](https://golang.org/doc/install).

### Disclaimer

**This project is only a proof of concept and SHOULDN'T be deployed on a production 
environment as it lacks even the most basic security measures.**

### Build the project

The project includes a Makefile, to build execute:

```bash
make
```

This should generate a `transcribe-server` binary.

### Running the server

The server can be configured using command line arguments or environment variables. Command line arguments take precedence over environment variables.

#### Environment Configuration

Environment variables can be set in several ways:

1. **Direct export**: Set variables in your shell session
2. **`.env` file**: Create a `.env` file in the project root (recommended for development)
3. **System environment**: Set variables at the system level

**Example `.env` file:**
```bash
# Copy env.example to .env and fill in your values
cp env.example .env

# Edit .env with your actual credentials
GOOGLE_CREDENTIALS=/path/to/your/google-credentials.json
AZURE_SPEECH_KEY=your_azure_subscription_key
AZURE_SPEECH_REGION=eastus
XUNFEI_APP_ID=your_xunfei_app_id
XUNFEI_API_KEY=your_xunfei_api_key
XUNFEI_API_SECRET=your_xunfei_api_secret
```

**Note:** The `env.example` file is included in the project for easy setup.

#### Command Line Arguments

```bash
# Basic usage with vendor selection
./webrtc-transcriber --vendor=whisper --model=base --output=./my_output

# Available options:
--vendor string     Transcription vendor: google, azure, baidu, xunfei, whisper, recorder (default "whisper")
--model string      Whisper model: tiny, base, small, medium, large (default "tiny")
--output string     Output directory for WAV and TXT files (default "recordings")
--language string   Source language (e.g., en, cn, auto) (default "auto")
--keep_wav          Keep generated WAV files (default: false)
--keep_txt          Keep generated TXT files (default: false)
--http.port string  HTTP listen port (default "9070")
--stun.server string STUN server URL (default "stun:stun.l.google.com:19302")

# Examples:
# Use Google Speech with credentials
export GOOGLE_CREDENTIALS=/path/to/credentials.json
./webrtc-transcriber --vendor=google

# Use Azure Speech Service
./webrtc-transcriber --vendor=azure

# Use Whisper with custom model and output (default vendor)
./webrtc-transcriber --model=base --output=./my_output

# Use Recorder to save WAV files
./webrtc-transcriber --vendor=recorder --output=./recordings

# Keep generated files
./webrtc-transcriber --keep_wav --keep_txt
```

#### Environment Variable Configuration

The server can also be configured to use different transcription services based on available credentials:

#### Option 1: Google Speech-to-Text (Cloud-based)
```bash
./webrtc-transcriber --google.cred=/path/to/google-credentials.json
```

**Requirements:**
- Google Cloud project with Speech-to-Text API enabled
- Service account credentials file

#### Option 2: Azure Speech Service (Enterprise-grade)
```bash
export AZURE_SPEECH_KEY="your_subscription_key"
export AZURE_SPEECH_REGION="your_region"
./webrtc-transcriber
```

**Requirements:**
- Azure account with Speech service resource
- Subscription key and region from Azure portal

#### Option 3: Baidu Speech Recognition (Chinese-optimized)
```bash
export BAIDU_APP_ID="your_app_id"
export BAIDU_API_KEY="your_api_key"
export BAIDU_SECRET_KEY="your_secret_key"
./webrtc-transcriber
```

**Requirements:**
- Baidu AI Platform account with Speech Recognition enabled
- App ID, API Key, and Secret Key

#### Option 4: Xunfei (讯飞) Speech Recognition (Chinese-optimized)
```bash
export XUNFEI_APP_ID="your_app_id"
export XUNFEI_API_KEY="your_api_key"
export XUNFEI_API_SECRET="your_api_secret"
export XUNFEI_API_URL="wss://iat-api.xfyun.cn/v2/iat"  # Optional, defaults to official endpoint
./webrtc-transcriber
```

**Requirements:**
- Xunfei Open Platform account
- Speech recognition service enabled
- **Optional**: Custom API endpoint via `XUNFEI_API_URL` (defaults to official endpoint if not set)

#### Option 5: Whisper Speech Recognition (Local/Offline)
```bash
# Option A: Auto-detection (recommended for whisper-ctranslate2)
./webrtc-transcriber

# Option B: Custom paths
export WHISPER_MODEL_PATH="/path/to/whisper/model"
export WHISPER_PATH="/path/to/whisper/executable"
./webrtc-transcriber
```

**Requirements:**
- Whisper executable (whisper-ctranslate2, whisper.cpp, etc.)
- Pre-trained Whisper model file (auto-downloaded on first use)
- Sufficient CPU/GPU resources

#### Option 6: Local Audio Recorder (WAV files)
```bash
# Use default output directory (./recordings)
./webrtc-transcriber

# Or specify custom output directory
export RECORDER_OUTPUT_DIR="/path/to/recordings"
./webrtc-transcriber
```

**Requirements:**
- No external services needed
- Sufficient disk space for audio recordings

#### Common Options

`--http.port` (Optional)
Specifies the port where the HTTP server should listen, by default the port 9070 is used.

`--stun.server` (Optional)
Allows to specify a different [STUN](https://es.wikipedia.org/wiki/STUN) server, by default a Google STUN server is used.

### Service Selection Priority

The server automatically selects the transcription service in this order:
1. **Google Speech** (if `--google.cred` is provided)
2. **Azure Speech** (if `AZURE_SPEECH_KEY` and `AZURE_SPEECH_REGION` are set)
3. **Baidu Speech** (if `BAIDU_*` environment variables are set)
4. **Xunfei** (if environment variables are set)
5. **Whisper** (if `WHISPER_MODEL_PATH` or `WHISPER_PATH` are set)
6. **Recorder** (fallback - no credentials needed)

### Demo page

The demo works on Chrome 75, Firefox 67 and Safari 12.1.1

![Demo screenshot](docs/demo.png)

To run the demo execute the server and navigate to `http://localhost:9070`. 

After pressing the **Start** button a dialog asking for permission to access the microphone should appear. 
After granting access a WebRTC connection is made to the local server, where audio data is decoded and streamed 
to the selected transcription service.

Say something and press the **Stop** button, the results (if any) should appear on screen.

### Transcription Services

#### Google Speech-to-Text
- **Best for**: English and multi-language support
- **Features**: High accuracy, multiple language models
- **Cost**: Pay-per-use cloud service
- **Setup**: Requires Google Cloud credentials

#### Azure Speech Service
- **Best for**: Enterprise applications, Windows ecosystem
- **Features**: High accuracy, 100+ languages, custom models, speaker identification
- **Cost**: Free tier (5 hours/month), then $16.00/hour
- **Setup**: Requires Azure Speech service subscription key and region

#### Baidu Speech Recognition
- **Best for**: Chinese language applications, cost-effective Chinese recognition
- **Features**: Excellent Chinese accuracy, multiple dialects, real-time streaming
- **Cost**: Free tier (500 requests/day), competitive paid pricing
- **Setup**: Requires Baidu AI Platform App ID, API Key, and Secret Key

#### Whisper Speech Recognition
- **Best for**: Privacy-conscious applications, offline use, high accuracy
- **Features**: Local processing, 100+ languages, no external dependencies
- **Cost**: Free (computational resources only)
- **Setup**: Requires Whisper executable and pre-trained model

#### Xunfei (讯飞) Speech Recognition
- **Best for**: Chinese language and dialects
- **Features**: 23+ Chinese dialect support, dynamic correction
- **Cost**: Often more affordable for Chinese
- **Setup**: Requires Xunfei platform credentials

#### Local Audio Recorder
- **Best for**: Audio archiving, debugging, offline processing
- **Features**: No external dependencies, WAV file output
- **Cost**: Free (local storage only)
- **Setup**: No credentials needed

### Documentation

For detailed setup instructions, see:
- [Azure Speech Setup Guide](docs/AZURE_SETUP.md)
- [Baidu Speech Setup Guide](docs/BAIDU_SETUP.md)
- [Whisper Setup Guide](docs/WHISPER_SETUP.md)
- [Xunfei Setup Guide](docs/XUNFEI_SETUP.md)
- [Recorder Setup Guide](docs/RECORDER_SETUP.md)

### Architecture

![Architecture and data flow](docs/architecture.png)



- Unit tests.
- Be able to specify the desired language.
- Support for interim results.

### License

MIT - see [LICENSE](LICENSE) for the full text.
