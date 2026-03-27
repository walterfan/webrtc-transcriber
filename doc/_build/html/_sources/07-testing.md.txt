# Testing Strategy

## Current State

:::{admonition} Gap
:class: warning

The project currently has **no automated tests**. The only test-like file is `test_wav.go`, which is a standalone manual test for the `RecorderTranscriber` -- it is not part of `go test`.
:::

## Recommended Test Pyramid

### Unit Tests (Target: 70% of test effort)

| Package | What to Test | Priority |
|---------|-------------|----------|
| `internal/transcribe` | `Result` serialization, `StreamOptions` defaults | High |
| `internal/transcribe` | Whisper command construction, WAV header generation | High |
| `internal/transcribe` | Each vendor's `CreateStreamWithOptions` with mock deps | Medium |
| `internal/rtc` | Opus decode correctness with known audio samples | Medium |
| `internal/session` | `newSessionRequest` JSON parsing, validation | Medium |
| `cmd/transcribe-server` | `selectVendor()` with various env/flag combinations | High |
| `cmd/transcribe-server` | `loadAccounts()` parsing | High |
| `cmd/transcribe-server` | `SessionStore` concurrency safety | Medium |

### Integration Tests (Target: 20% of test effort)

| Scope | What to Test |
|-------|-------------|
| HTTP API | `/login`, `/logout`, `/auth/status` with valid and invalid credentials |
| HTTP API | `/files` returns correct file listing |
| HTTP API | `/delete/<file>` with path traversal attempts |
| Session flow | Full SDP exchange with a mock RTC service |
| Whisper integration | End-to-end: WAV input → CLI invocation → text output |

### End-to-End Tests (Target: 10% of test effort)

| Scope | What to Test |
|-------|-------------|
| Browser → Server | WebRTC connection setup, audio streaming, transcript display |
| File management | Upload (record), list, play, delete through the UI |

## Critical Path Test Checklist

- [ ] Login with valid credentials returns 200 and sets cookie
- [ ] Login with invalid credentials returns 401
- [ ] Unauthenticated requests to protected routes return 401
- [ ] Expired session returns 401
- [ ] SDP exchange returns a valid SDP answer
- [ ] Opus packets decode to valid PCM
- [ ] Whisper subprocess invocation produces text output
- [ ] Whisper fallback to Recorder works when whisper-ctranslate2 is absent
- [ ] File listing returns all files sorted by modification time
- [ ] File deletion removes the file and returns success
- [ ] Directory traversal in delete is blocked

## Test Data

- **Audio fixtures**: Pre-recorded Opus and WAV files with known transcription text.
- **SDP fixtures**: Known-good SDP offer/answer pairs for session tests.
- **Environment fixtures**: `.env.test` files with mock credentials for vendor selection tests.

## Coverage Thresholds

| Package | Target |
|---------|--------|
| `internal/transcribe` | 80% |
| `internal/rtc` | 60% |
| `internal/session` | 80% |
| `cmd/transcribe-server` | 70% |

## Running Tests

```bash
# Run all tests
go test ./...

# Run with verbose output
go test -v ./...

# Run with coverage
go test -cover ./...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## Flaky Test Handling

- WebRTC-related tests may be timing-sensitive. Use generous timeouts and retry mechanisms.
- Whisper subprocess tests depend on external binary availability. Skip gracefully with `t.Skip()` when `whisper-ctranslate2` is not installed.
- File-based tests should use `t.TempDir()` for isolation.
