# Architecture

## Component Overview

```
┌─────────────────────────────────────────────────────────────────────┐
│                          Browser (Vue 3)                            │
│  ┌────────┐  ┌──────────┐  ┌────────────┐  ┌───────────────────┐  │
│  │ Login  │  │ WebRTC   │  │ Audio Viz  │  │ File Management   │  │
│  │ Form   │  │ Client   │  │ (Waveform) │  │ (List/Play/Delete)│  │
│  └────────┘  └────┬─────┘  └────────────┘  └───────────────────┘  │
│                    │ SDP Offer (HTTP POST)                          │
│                    │ Opus Audio (RTP)                               │
│                    │ Results (DataChannel)                          │
└────────────────────┼───────────────────────────────────────────────┘
                     │
     ════════════════╪═══════════════════  Network
                     │
┌────────────────────┼───────────────────────────────────────────────┐
│                    ▼       Go Server (single process)              │
│  ┌──────────────────────────────────────────────────────────────┐  │
│  │  cmd/transcribe-server/main.go                               │  │
│  │  HTTP Mux + Auth Middleware                                  │  │
│  │  Routes: /login /logout /auth/status /session /files /delete │  │
│  └──────────┬───────────────────────────────────────────────────┘  │
│             │                                                      │
│  ┌──────────▼──────────┐                                          │
│  │  session.Handler    │  POST /session                           │
│  │  (SDP Exchange)     │  Accept SDP offer → return SDP answer    │
│  └──────────┬──────────┘                                          │
│             │                                                      │
│  ┌──────────▼──────────┐     ┌────────────────────────┐           │
│  │  rtc.PionService    │────►│  opus.Decoder           │           │
│  │  (Pion WebRTC)      │     │  Opus RTP → PCM 16kHz   │           │
│  │  PeerConnection     │     └────────────┬───────────┘           │
│  │  DataChannel        │                  │ PCM bytes             │
│  └──────────┬──────────┘                  │                       │
│             │                  ┌───────────▼───────────┐           │
│             │                  │  transcribe.Service   │           │
│             │                  │  ┌─────────────────┐  │           │
│             │                  │  │ WhisperTranscr. │  │ default   │
│             │                  │  │ GoogleSpeech    │  │           │
│             │                  │  │ AzureTranscr.   │  │           │
│             │                  │  │ BaiduTranscr.   │  │           │
│             │                  │  │ IflyTekTranscr. │  │           │
│             │                  │  │ RecorderTranscr.│  │           │
│             │                  │  └─────────────────┘  │           │
│             │  Result channel  └───────────┬───────────┘           │
│             ◄──────────────────────────────┘                       │
│             │ Send via DataChannel                                 │
│             ▼                                                      │
│         Browser                                                    │
└────────────────────────────────────────────────────────────────────┘
```

## Key Call Chains

### Use Case 1: Real-time Transcription (Full Mode)

```
Browser                  Go Server
  │                        │
  │  POST /login           │
  │───────────────────────►│  loginHandler() validates credentials
  │  Set-Cookie            │  sessionStore.createSession()
  │◄───────────────────────│
  │                        │
  │  POST /session         │
  │  {offer, language}     │
  │───────────────────────►│  session.Handler
  │                        │    → rtc.CreatePeerConnectionWithOptions()
  │                        │    → pion.NewPeerConnection()
  │                        │    → transcribe.CreateStreamWithOptions()
  │  {answer}              │    → peer.ProcessOffer(offer)
  │◄───────────────────────│
  │                        │
  │  Opus RTP packets ════►│  rtc.onTrack callback
  │                        │    → opus.Decode() → PCM
  │                        │    → stream.Write(pcm)
  │                        │    → (Whisper writes WAV, invokes CLI)
  │                        │
  │  ◄════ DataChannel ═══ │  stream.Results() → Result{Text, Confidence}
  │  display transcript    │    → dataChannel.Send(json)
  │                        │
  │  ICE disconnect        │
  │───────────────────────►│  peer.Close() → stream.Close()
```

### Use Case 2: File Management

```
Browser                  Go Server
  │  GET /files            │
  │───────────────────────►│  os.ReadDir(output) → JSON list
  │  [{name, modTime}]     │
  │◄───────────────────────│
  │                        │
  │  GET /recordings/x.wav │
  │───────────────────────►│  http.FileServer serves the file
  │  audio data            │
  │◄───────────────────────│
  │                        │
  │  DELETE /delete/x.wav  │
  │───────────────────────►│  os.Remove() → JSON response
  │  {success: true}       │
  │◄───────────────────────│
```

## Module Dependencies

```
cmd/transcribe-server
  ├── internal/rtc          (creates PionRtcService)
  ├── internal/session      (creates HTTP handler)
  └── internal/transcribe   (vendor selection)

internal/session
  └── internal/rtc          (uses rtc.Service interface)

internal/rtc
  └── internal/transcribe   (uses transcribe.Service/Stream)

internal/transcribe
  └── (no internal deps)    (leaf package)
```

**Rule**: `transcribe` is the leaf package. It must not import from `rtc`, `session`, or `cmd`.

## Cross-Cutting Concerns

### Authentication

- Cookie-based sessions with random 32-byte hex tokens.
- In-memory session store (`sync.RWMutex`-protected map).
- 24-hour session duration.
- Accounts loaded from `accounts` env var (format: `alice:pass,bob:pass`).
- Public routes: `/login`, `/logout`, `/auth/status`.
- All other routes pass through `authMiddleware`.

### Error Handling

- Go standard `error` return values; no custom error types.
- HTTP handlers return appropriate status codes (400, 401, 405, 500).
- `log.Fatalf` for fatal startup errors; `log.Printf` for warnings.
- Whisper failure gracefully falls back to Recorder service.

### Configuration

- Environment variables via `.env` file (loaded by godotenv).
- CLI flags for runtime options (`--vendor`, `--model`, `--language`, `--output`, `--http.port`, `--stun.server`).
- CLI flags take precedence over environment variables.
- No configuration file format (YAML/TOML/JSON); env vars and flags only.

### Logging

- Go standard `log` package.
- No structured logging, no log levels beyond `Printf` / `Fatalf`.
- Key events logged: account loading, vendor selection, session creation, file operations.

### Concurrency

- `sync.RWMutex` for session store.
- Each WebRTC peer connection runs in its own goroutine.
- Transcription results delivered via Go channels (`<-chan Result`).
- Graceful shutdown via OS signal interception.
