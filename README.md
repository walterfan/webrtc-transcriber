# 🎙️ Lazy Speech To Text Converter

<p align="center">
  <img src="docs/snapshot.png" alt="Demo Screenshot" width="600">
</p>

<p align="center">
  <strong>Transform speech to text effortlessly with WebRTC and AI</strong>
</p>

<p align="center">
  <a href="#-quick-start">Quick Start</a> •
  <a href="#-features">Features</a> •
  <a href="#-web-interface">Web Interface</a> •
  <a href="#-transcription-services">Services</a> •
  <a href="#-configuration">Configuration</a>
</p>

---

## ✨ Features

| Feature | Description |
|---------|-------------|
| 🎤 **Real-time Streaming** | WebRTC-based audio capture with low latency |
| 🌍 **99+ Languages** | Powered by Whisper AI - works offline |
| 🔒 **Privacy First** | Local processing - your audio never leaves your machine |
| 📱 **Cross-platform** | Works on Chrome, Firefox, and Safari |
| 🎛️ **Flexible Options** | Record only, transcribe only, or both |
| 📊 **Visual Feedback** | Real-time audio waveform visualization |
| 🔐 **User Authentication** | Simple login system for access control |

---

## 🚀 Quick Start

Get up and running in **under 2 minutes**:

```bash
# 1. Install Whisper (one-time)
pip install whisper-ctranslate2

# 2. Build the project
make

# 3. Run the server
./webrtc-transcriber

# 4. Open browser
open http://localhost:9070
```

**That's it!** No cloud accounts, no API keys, no configuration needed.

---

## 🖥️ Web Interface

<table>
<tr>
<td width="50%">

### Controls
- **🎤 Record Audio** - Capture audio from your microphone
- **📝 Transcribe Audio** - Convert speech to text
- **🌐 Language Selection** - Choose from 20+ languages
- **🎧 Device Selection** - Pick your audio input device

</td>
<td width="50%">

### Features
- **📈 Live Waveform** - See your audio in real-time
- **📁 File Management** - Play, download, delete files
- **👁️ Preview** - View transcription results instantly
- **📊 Stats** - Recording duration, codec info

</td>
</tr>
</table>

### Operation Modes

| Mode | Record | Transcribe | Description |
|------|:------:|:----------:|-------------|
| **Full** | ✅ | ✅ | Record and transcribe in one step (default) |
| **Record Only** | ✅ | ❌ | Save audio for later transcription |
| **Transcribe Only** | ❌ | ✅ | Transcribe existing recordings |

---

## 🎯 Transcription Services

### ⭐ Whisper (Default & Recommended)

**Best for:** Privacy, offline use, high accuracy

```bash
# Just run - Whisper is the default!
./webrtc-transcriber
```

| Model | Size | Speed | Accuracy | Use Case |
|-------|------|-------|----------|----------|
| tiny | 39MB | ⚡⚡⚡⚡ | ★★☆☆☆ | Quick tests |
| base | 74MB | ⚡⚡⚡ | ★★★☆☆ | General use |
| **small** | 244MB | ⚡⚡ | ★★★★☆ | **Default - Best balance** |
| medium | 769MB | ⚡ | ★★★★★ | High accuracy |
| large | 1.5GB | 🐢 | ★★★★★ | Maximum accuracy |

> 💡 Models auto-download to `~/.cache/whisper/` on first use

### Other Services

<details>
<summary><b>☁️ Google Speech-to-Text</b></summary>

```bash
export GOOGLE_CREDENTIALS=/path/to/credentials.json
./webrtc-transcriber --vendor=google
```
- 125+ languages
- High accuracy
- Pay-per-use

</details>

<details>
<summary><b>🔷 Azure Speech Service</b></summary>

```bash
export AZURE_SPEECH_KEY="your_key"
export AZURE_SPEECH_REGION="eastus"
./webrtc-transcriber --vendor=azure
```
- Enterprise-grade
- 100+ languages
- Free tier available

</details>

<details>
<summary><b>🇨🇳 Baidu Speech (Chinese)</b></summary>

```bash
export BAIDU_APP_ID="your_app_id"
export BAIDU_API_KEY="your_api_key"
export BAIDU_SECRET_KEY="your_secret_key"
./webrtc-transcriber --vendor=baidu
```
- Optimized for Chinese
- Multiple dialects

</details>

<details>
<summary><b>🇨🇳 Xunfei/讯飞 (Chinese)</b></summary>

```bash
export XUNFEI_APP_ID="your_app_id"
export XUNFEI_API_KEY="your_api_key"
export XUNFEI_API_SECRET="your_api_secret"
./webrtc-transcriber --vendor=xunfei
```
- 23+ Chinese dialects
- Real-time streaming

</details>

<details>
<summary><b>💾 Local Recorder (WAV only)</b></summary>

```bash
./webrtc-transcriber --vendor=recorder --output=./recordings
```
- No transcription
- Just saves audio files

</details>

---

## ⚙️ Configuration

### Command Line Options

```bash
./webrtc-transcriber [options]

Options:
  --vendor string     Service: whisper, google, azure, baidu, xunfei, recorder
                      (default "whisper")
  --model string      Whisper model: tiny, base, small, medium, large
                      (default "small")
  --language string   Language code: en, zh, ja, auto, etc.
                      (default "auto")
  --output string     Output directory for files
                      (default "recordings")
  --keep_wav          Keep WAV files after transcription
  --keep_txt          Keep TXT files
  --http.port string  HTTP server port (default "9070")
```

### Environment Variables

Create a `.env` file in the project root:

```bash
# Authentication (required)
accounts=alice:password123,bob:secret456

# Cloud services (optional)
GOOGLE_CREDENTIALS=/path/to/credentials.json
AZURE_SPEECH_KEY=your_azure_key
AZURE_SPEECH_REGION=eastus
```

---

## 🌍 Supported Languages

<table>
<tr>
<td>

| Language | Code |
|----------|------|
| English | `en` |
| Chinese | `zh` |
| Japanese | `ja` |
| Korean | `ko` |
| Spanish | `es` |

</td>
<td>

| Language | Code |
|----------|------|
| French | `fr` |
| German | `de` |
| Italian | `it` |
| Portuguese | `pt` |
| Russian | `ru` |

</td>
<td>

| Language | Code |
|----------|------|
| Arabic | `ar` |
| Hindi | `hi` |
| Thai | `th` |
| Vietnamese | `vi` |
| Auto Detect | `auto` |

</td>
</tr>
</table>

> 💡 For Chinese, use `small` or `medium` model for best results

---

## 🏗️ Architecture

```
┌─────────────┐     WebRTC      ┌─────────────────┐     Audio      ┌──────────────┐
│   Browser   │ ◄─────────────► │  Go Server      │ ─────────────► │  Whisper AI  │
│  (Web UI)   │                 │  (Pion WebRTC)  │                │  (or Cloud)  │
└─────────────┘                 └─────────────────┘                └──────────────┘
       │                               │                                  │
       │    DataChannel               │                                  │
       ◄──────────────────────────────┼──────────────────────────────────┘
              (Transcription Results)
```

---

## 🛠️ Tech Stack

<table>
<tr>
<td align="center" width="20%">

### Backend
<img src="https://cdn.jsdelivr.net/gh/devicons/devicon/icons/go/go-original-wordmark.svg" width="60" height="60"/>

**Go 1.12+**

</td>
<td align="center" width="20%">

### WebRTC
<img src="https://webrtc.github.io/webrtc-org/assets/images/webrtc-logo-vert-retro-255x305.png" width="60" height="60"/>

**Pion WebRTC**

</td>
<td align="center" width="20%">

### Frontend
<img src="https://cdn.jsdelivr.net/gh/devicons/devicon/icons/react/react-original.svg" width="60" height="60"/>

**React 18**

</td>
<td align="center" width="20%">

### AI/ML
<img src="https://cdn.jsdelivr.net/gh/devicons/devicon/icons/python/python-original.svg" width="60" height="60"/>

**Whisper AI**

</td>
<td align="center" width="20%">

### Styling
<img src="https://bulma.io/assets/images/bulma-logo.png" width="60" height="60"/>

**Bulma CSS**

</td>
</tr>
</table>

### Full Stack Overview

| Layer | Technology | Purpose |
|-------|------------|---------|
| **Frontend** | React 18 (via CDN) | UI components, state management |
| **Styling** | Bulma CSS + Custom CSS | Responsive design, modern UI |
| **Icons** | Font Awesome 6 | UI icons and visual elements |
| **Audio API** | Web Audio API | Waveform visualization, audio processing |
| **Real-time** | WebRTC (Pion) | Low-latency audio streaming |
| **Data Channel** | WebRTC DataChannel | Transcription results delivery |
| **Backend** | Go (Golang) | HTTP server, WebRTC signaling |
| **Audio Codec** | Opus | High-quality audio compression |
| **Transcription** | Whisper (ctranslate2) | Speech-to-text AI model |
| **Session** | Cookie-based auth | User authentication |
| **Config** | godotenv | Environment variable management |

### Key Libraries & Dependencies

```
Backend (Go):
├── github.com/pion/webrtc/v2     # WebRTC implementation
├── github.com/gorilla/websocket  # WebSocket support
├── github.com/joho/godotenv      # .env file loading
└── gopkg.in/hraban/opus.v2       # Opus audio codec

Frontend (Browser):
├── React 18                       # UI framework
├── Bulma 0.9.4                   # CSS framework
├── Font Awesome 6                # Icons
└── WebRTC Adapter                # Browser compatibility

AI/ML (Python):
└── whisper-ctranslate2           # Fast Whisper implementation
```

---

## 📋 Requirements

| Component | Version | Required |
|-----------|---------|:--------:|
| Go | 1.12+ | ✅ |
| Python | 3.8+ | ⚠️ (for Whisper) |
| Chrome | 75+ | ✅ |
| Firefox | 67+ | ✅ |
| Safari | 12.1+ | ✅ |

> ⚠️ Python is only required if using Whisper (default). Cloud services don't need Python.

---

## 🔧 Development

```bash
# Clone the repository
git clone https://github.com/walterfan/webrtc-transcriber.git
cd webrtc-transcriber

# Install Go dependencies
go mod download

# Install Whisper (for transcription)
pip install whisper-ctranslate2

# Build
make

# Run the server
./webrtc-transcriber

# Or run directly (development)
go run ./cmd/transcribe-server/main.go
```

### Project Structure

```
webrtc-transcriber/
├── cmd/
│   └── transcribe-server/
│       └── main.go           # Application entry point
├── internal/
│   ├── rtc/
│   │   ├── pion.go          # WebRTC implementation (Pion)
│   │   └── service.go       # RTC service interface
│   ├── session/
│   │   ├── handler.go       # HTTP session handler
│   │   └── payload.go       # Request/response types
│   └── transcribe/
│       ├── service.go       # Transcription interface
│       ├── whisper.go       # Whisper implementation
│       ├── gspeech.go       # Google Speech implementation
│       ├── azure.go         # Azure Speech implementation
│       ├── baidu.go         # Baidu Speech implementation
│       ├── iflytek.go       # Xunfei implementation
│       └── recorder.go      # Local recorder implementation
├── web/
│   ├── index.html           # Main HTML page
│   ├── js/
│   │   └── app.js           # React application
│   └── vendor/              # Local CSS/JS libraries
├── docs/                    # Documentation
├── recordings/              # Output directory (default)
├── .env                     # Environment configuration
├── Makefile                 # Build configuration
└── README.md
```

---

## 📚 Documentation

- [Whisper Setup Guide](docs/WHISPER_SETUP.md)
- [Azure Speech Setup](docs/AZURE_SETUP.md)
- [Baidu Speech Setup](docs/BAIDU_SETUP.md)
- [Xunfei Setup Guide](docs/XUNFEI_SETUP.md)

---

## ⚠️ Disclaimer

This project is a **proof of concept** and should not be deployed in production without implementing proper security measures.

---

## 📄 License

MIT - see [LICENSE](LICENSE) for details.

---

<p align="center">
  Made with ❤️ by <a href="mailto:walter.fan@gmail.com">Walter Fan</a>
</p>
