# Business Workflows

## Workflow: Real-time Speech Transcription

### Overview

The primary workflow captures audio from the browser microphone, streams it to the server via WebRTC, transcribes it using an AI engine, and returns results to the browser in real time.

### Steps

1. **User authenticates**: POST `/login` with username/password. Server sets session cookie.
2. **User selects options**: Language, operation mode (Full / Record Only / Transcribe Only), audio device.
3. **Browser creates SDP offer**: `useWebRTC.ts` creates an RTCPeerConnection with an audio track and a DataChannel.
4. **SDP exchange**: Browser sends the offer to `POST /session`. Server creates a Pion peer connection, processes the offer, returns an SDP answer.
5. **ICE connectivity**: Browser and server exchange ICE candidates. STUN server resolves NAT traversal.
6. **Audio streaming**: Browser sends Opus-encoded audio via RTP. Server receives it in `onTrack`.
7. **Opus decoding**: `opus.Decode()` converts Opus packets to 16 kHz PCM samples.
8. **Transcription**: PCM data is written to the `transcribe.Stream`. For Whisper, this accumulates into a WAV buffer and periodically invokes the `whisper-ctranslate2` CLI.
9. **Result delivery**: The stream's `Results()` channel emits `Result` structs. The RTC layer serializes them to JSON and sends them over the DataChannel.
10. **Display**: The Vue frontend receives DataChannel messages and renders the transcript text.
11. **Session end**: User clicks stop or disconnects. Peer connection closes, stream finalizes.

### Input/Output

- **Input**: Opus audio packets (48 kHz, mono) from browser microphone.
- **Output**: JSON `Result` objects `{text, confidence, final, audio_file, text_file}` over DataChannel.

### Key Validations

- Session cookie must be valid and not expired.
- SDP offer must be a valid WebRTC offer.
- Audio track must be present in the peer connection.

### Error Branches

- Invalid credentials: HTTP 401 returned, no session created.
- Whisper not installed: Automatic fallback to Recorder (WAV-only, no transcription).
- ICE connection failure: Peer connection times out; browser reports error.
- Transcription subprocess crash: Error logged, stream may stop producing results.

### Data Entities

- `transcribe.Result`: Text, Confidence, Final, AudioFile, TextFile.
- `session.newSessionRequest`: SDP Offer, Language, Transcribe flag.
- `session.newSessionResponse`: SDP Answer.
- WAV files: Written to `recordings/` directory.
- TXT files: Transcription text saved alongside WAV.

### Code Entry Points

| Step | File | Function/Method |
|------|------|-----------------|
| Login | `cmd/transcribe-server/main.go` | `loginHandler()` |
| SDP Exchange | `internal/session/handler.go` | `MakeHandler()` |
| Peer Connection | `internal/rtc/pion.go` | `NewPionRtcService()`, `CreatePeerConnectionWithOptions()` |
| Opus Decode | `internal/rtc/opus.go` | `Decode()` |
| Whisper Transcribe | `internal/transcribe/whisper.go` | `NewWhisperTranscriber()`, `Write()`, `Close()` |
| Recorder (fallback) | `internal/transcribe/recorder.go` | `NewRecorderTranscriber()` |
| Frontend WebRTC | `frontend/src/composables/useWebRTC.ts` | `startSession()` |

---

## Workflow: File Management

### Overview

Authenticated users can list, play, download, and delete recordings from the web UI.

### Steps

1. **List files**: `GET /files` returns a JSON array of `{name, modTime}` sorted by modification time (newest first).
2. **Play/Download**: `GET /recordings/<filename>` serves the file via Go's `http.FileServer`.
3. **Delete**: `DELETE /delete/<filename>` removes the file from disk.

### Validations

- All requests require a valid session cookie.
- Delete sanitizes the filename to prevent directory traversal (`..`, `/`, `\` stripped).
- Only `DELETE` method accepted for deletion.

### Code Entry Points

| Step | File | Function |
|------|------|----------|
| List | `cmd/transcribe-server/main.go` | `/files` handler (inline) |
| Serve | `cmd/transcribe-server/main.go` | `/recordings/` FileServer |
| Delete | `cmd/transcribe-server/main.go` | `/delete/` handler (inline) |
| Frontend | `frontend/src/composables/useFileManager.ts` | `fetchFiles()`, `deleteFile()` |

---

## Workflow: Vendor Selection

### Overview

At startup, the server selects a transcription vendor based on CLI flags and environment variables.

### Priority Order

1. `--vendor` CLI flag (explicit selection).
2. `GOOGLE_CREDENTIALS` environment variable (auto-detect Google).
3. `AZURE_SPEECH_KEY` + `AZURE_SPEECH_REGION` (auto-detect Azure).
4. `BAIDU_APP_ID` + `BAIDU_API_KEY` + `BAIDU_SECRET_KEY` (auto-detect Baidu).
5. `XUNFEI_APP_ID` + `XUNFEI_API_KEY` + `XUNFEI_API_SECRET` (auto-detect Xunfei).
6. Whisper (auto-detect via PATH).
7. Recorder (WAV-only fallback, always available).

### Code Entry Point

- `cmd/transcribe-server/main.go`: `selectVendor()`
