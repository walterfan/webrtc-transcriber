# Data Model and API Contracts

## Data Model

This project has no database. All state is in-memory or on the filesystem.

### In-Memory State

| Entity | Location | Lifetime |
|--------|----------|----------|
| `SessionStore` | `cmd/transcribe-server/main.go` | Process lifetime |
| `SessionData{Username, ExpiresAt}` | Map in SessionStore | 24 hours per session |
| `accounts` | Map in main.go | Process lifetime (loaded from env at startup) |

### Filesystem State

| Artifact | Directory | Format | Naming Convention |
|----------|-----------|--------|-------------------|
| Audio recordings | `recordings/` (configurable) | WAV, 16 kHz, mono, 16-bit PCM | `whisper_audio_<n>_<YYYYMMDD>_<HHMMSS>.wav` |
| Transcription text | `recordings/` (configurable) | Plain text (UTF-8) | `whisper_audio_<n>_<YYYYMMDD>_<HHMMSS>.txt` |

## HTTP API

### Public Endpoints

#### `POST /login`

Authenticate a user and create a session.

- **Content-Type**: `application/x-www-form-urlencoded`
- **Body**: `username=<user>&password=<pass>`
- **Success (200)**:
  ```json
  {"success": true, "username": "alice"}
  ```
  Sets `session_token` cookie (HttpOnly, SameSite=Strict, 24h TTL).
- **Failure (401)**:
  ```json
  {"success": false, "message": "Invalid username or password"}
  ```

#### `POST /logout`

Destroy the current session.

- **Success (200)**:
  ```json
  {"success": true}
  ```
  Clears the `session_token` cookie.

#### `GET /auth/status`

Check current authentication state.

- **Authenticated**:
  ```json
  {"authenticated": true, "username": "alice"}
  ```
- **Not authenticated**:
  ```json
  {"authenticated": false}
  ```

### Protected Endpoints

All protected endpoints require a valid `session_token` cookie. Unauthorized requests receive HTTP 401.

#### `POST /session`

Establish a WebRTC peer connection via SDP exchange.

- **Request**:
  ```json
  {
    "offer": "<SDP offer string>",
    "language": "en",
    "transcribe": true
  }
  ```
- **Response**:
  ```json
  {
    "answer": "<SDP answer string>"
  }
  ```

#### `GET /files`

List all files in the recordings directory, sorted by modification time (newest first).

- **Response**:
  ```json
  [
    {"name": "whisper_audio_1_20251219_220725.wav", "modTime": 1734645645000},
    {"name": "whisper_audio_1_20251219_220725.txt", "modTime": 1734645645000}
  ]
  ```

#### `GET /recordings/<filename>`

Serve a recording file. Handled by Go `http.FileServer`.

#### `DELETE /delete/<filename>`

Delete a recording file.

- **Success (200)**:
  ```json
  {"success": true}
  ```
- **Not Found (404)**:
  ```json
  {"success": false, "message": "File not found"}
  ```

## WebRTC DataChannel Messages

Transcription results are sent from server to browser as JSON over the WebRTC DataChannel.

```json
{
  "text": "Hello world",
  "confidence": 0.95,
  "final": true,
  "audio_file": "whisper_audio_1_20251219_220725.wav",
  "text_file": "whisper_audio_1_20251219_220725.txt"
}
```

| Field | Type | Description |
|-------|------|-------------|
| `text` | string | Transcribed text |
| `confidence` | float32 | Confidence score (0.0 - 1.0) |
| `final` | bool | Whether this is the final result for a segment |
| `audio_file` | string | Filename of the saved WAV file (if applicable) |
| `text_file` | string | Filename of the saved TXT file (if applicable) |

## Internal Interfaces

### `transcribe.Service`

```go
type Service interface {
    CreateStream() (Stream, error)
    CreateStreamWithOptions(opts StreamOptions) (Stream, error)
}
```

### `transcribe.Stream`

```go
type Stream interface {
    io.Writer                    // Write(pcm []byte) (n int, err error)
    io.Closer                    // Close() error
    Results() <-chan Result      // Asynchronous result delivery
}
```

### `rtc.Service`

```go
type Service interface {
    CreatePeerConnection() (PeerConnection, error)
    CreatePeerConnectionWithOptions(opts PeerConnectionOptions) (PeerConnection, error)
}
```

### `rtc.PeerConnection`

```go
type PeerConnection interface {
    io.Closer
    ProcessOffer(offer string) (string, error)
}
```
