# Repository Map

## Directory Structure

```
webrtc-transcriber/
├── cmd/
│   └── transcribe-server/
│       └── main.go                 # Application entry point, HTTP routing, auth
├── internal/
│   ├── rtc/
│   │   ├── service.go              # RTC Service and PeerConnection interfaces
│   │   ├── pion.go                 # Pion WebRTC implementation
│   │   └── opus.go                 # Opus-to-PCM decoder
│   ├── session/
│   │   ├── handler.go              # HTTP handler for /session (SDP exchange)
│   │   └── payload.go              # Request/response JSON types
│   └── transcribe/
│       ├── service.go              # Service and Stream interfaces, Result type
│       ├── whisper.go              # Whisper (ctranslate2) implementation
│       ├── gspeech.go              # Google Speech-to-Text
│       ├── azure.go                # Azure Speech Service
│       ├── baidu.go                # Baidu Speech Recognition
│       ├── iflytek.go              # Xunfei (讯飞) Speech Recognition
│       └── recorder.go            # WAV-only recorder (no transcription)
├── frontend/                       # Vue 3 + Vite + Tailwind frontend
│   ├── src/
│   │   ├── main.ts                 # Vue app bootstrap
│   │   ├── App.vue                 # Root component
│   │   ├── components/             # LoginForm, Navbar, Footer, Waveform, FileTable
│   │   └── composables/            # useAuth, useWebRTC, useFileManager, useAudioVisualization
│   ├── index.html
│   ├── package.json
│   ├── vite.config.ts
│   └── tailwind.config.js
├── web/                            # Legacy React/Bulma UI (superseded by frontend/)
├── docs/                           # Vendor-specific setup guides
│   ├── WHISPER_SETUP.md
│   ├── AZURE_SETUP.md
│   ├── BAIDU_SETUP.md
│   ├── XUNFEI_SETUP.md
│   └── RECORDER_SETUP.md
├── doc/                            # Project Knowledge Base (this documentation)
├── openspec/                       # OpenSpec change proposals
├── recordings/                     # Default output directory for WAV/TXT
├── go.mod / go.sum                 # Go module definition
├── Makefile                        # Build targets
├── env.example                     # Environment variable template
└── README.md                       # Project overview
```

## Directory Responsibilities

| Directory | Purpose |
|-----------|---------|
| `cmd/transcribe-server/` | Application entry point. CLI flag parsing, HTTP server setup, authentication, vendor selection. |
| `internal/rtc/` | WebRTC layer. Manages Pion peer connections, ICE/STUN, Opus decoding, and DataChannel communication. |
| `internal/session/` | HTTP session handler. Accepts SDP offers and returns SDP answers to establish WebRTC connections. |
| `internal/transcribe/` | Transcription abstraction layer. Defines the `Service`/`Stream` interfaces and all vendor implementations. |
| `frontend/` | Vue 3 single-page application. Handles UI, WebRTC client, auth, file management, and audio visualization. |
| `web/` | **Legacy** React + Bulma UI. Superseded by `frontend/` but kept for reference. |
| `docs/` | Vendor-specific setup guides (Whisper, Azure, Baidu, Xunfei). |
| `doc/` | Sphinx + MyST project knowledge base (this documentation). |
| `openspec/` | Change proposals and project conventions following the OpenSpec workflow. |
| `recordings/` | Default output directory for generated WAV and TXT files. |

## Key Entry Points

| Entry Point | File | Purpose |
|-------------|------|---------|
| Main | `cmd/transcribe-server/main.go` | CLI parsing, HTTP server, auth middleware, vendor selection |
| RTC Service | `internal/rtc/pion.go` | `NewPionRtcService()` creates the WebRTC service |
| Session Handler | `internal/session/handler.go` | `MakeHandler()` returns the `/session` HTTP handler |
| Transcription Interface | `internal/transcribe/service.go` | `Service` and `Stream` interfaces |
| Whisper Transcriber | `internal/transcribe/whisper.go` | `NewWhisperTranscriber()` creates the default vendor |
| Frontend App | `frontend/src/main.ts` | Vue 3 application bootstrap |
| WebRTC Client | `frontend/src/composables/useWebRTC.ts` | Browser-side WebRTC connection management |

## Conventions

### Naming

- **Go packages**: Lowercase, single-word (`rtc`, `session`, `transcribe`).
- **Go files**: Snake-case for multi-word names (`service.go`, `gspeech.go`).
- **Go interfaces**: Named by role (`Service`, `Stream`, `PeerConnection`).
- **Vue components**: PascalCase filenames (`LoginForm.vue`, `FileTable.vue`).
- **Composables**: `use` prefix (`useAuth.ts`, `useWebRTC.ts`).

### Layering

```
cmd/         (application wiring, CLI, HTTP routes)
  └─► internal/session    (HTTP handler, JSON serialization)
        └─► internal/rtc   (WebRTC peer connections, media)
              └─► internal/transcribe  (AI transcription backends)
```

Dependencies flow strictly downward. The `transcribe` package has no knowledge of `rtc` or `session`.

### Common Patterns

- **Strategy Pattern**: `transcribe.Service` interface with multiple implementations (Whisper, Google, Azure, Baidu, Xunfei, Recorder). Vendor selection happens once at startup in `selectVendor()`.
- **Streaming via channels**: `Stream.Results()` returns a `<-chan Result` for asynchronous delivery of transcription results.
- **Cookie-based auth**: In-memory session store with random tokens. No external session backend.
- **Graceful shutdown**: Signal handling with `os.Signal` and channel-based error propagation.
