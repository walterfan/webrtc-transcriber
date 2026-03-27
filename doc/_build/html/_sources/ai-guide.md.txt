# AI Usage Guide

This guide helps AI assistants understand and contribute to the WebRTC Transcriber project effectively.

## Orientation

1. **Start with** {doc}`00-overview` to understand the project purpose and stack.
2. **Navigate with** {doc}`01-repo-map` to find code locations.
3. **Understand flow with** {doc}`02-architecture` and {doc}`03-workflows`.
4. **Build and run with** {doc}`06-runbook`.

## Key Interfaces

The project follows a Strategy pattern for transcription vendors. All implementations satisfy:

```go
// internal/transcribe/service.go
type Service interface {
    CreateStream() (Stream, error)
    CreateStreamWithOptions(opts StreamOptions) (Stream, error)
}

type Stream interface {
    io.Writer
    io.Closer
    Results() <-chan Result
}
```

To add a new transcription vendor:

1. Create `internal/transcribe/newvendor.go`.
2. Implement `Service` and `Stream`.
3. Add a case to `selectVendor()` in `cmd/transcribe-server/main.go`.
4. Add env vars to `env.example`.

## Common Tasks

### Adding an HTTP endpoint

1. Define the handler function or use `session.MakeHandler()` pattern.
2. Register in the `mux` in `cmd/transcribe-server/main.go`.
3. If protected, wrap with `authMiddleware`.
4. Update the Vite proxy in `frontend/vite.config.ts` for dev mode.

### Modifying the frontend

1. Components live in `frontend/src/components/`.
2. State and logic live in `frontend/src/composables/`.
3. Run `cd frontend && npm run dev` for hot-reload development.
4. Build with `make build-frontend` before deploying.

### Debugging transcription issues

1. Check the vendor selection log at startup.
2. Enable verbose logging by watching stderr.
3. Check `recordings/` for generated WAV files.
4. Test Whisper directly: `whisper-ctranslate2 recordings/test.wav --model small`.

## Known Gaps

- `/transcribe` endpoint referenced in `useFileManager.ts` is not implemented on the backend.
- No automated tests exist. `test_wav.go` is a manual utility.
- `web/` directory is a legacy UI; all development should target `frontend/`.
- Go module declares `go 1.12` which is very old; may need updating for newer toolchains.

## Coding Guidelines

- Follow existing patterns: strategy interfaces in `transcribe`, Pion usage in `rtc`.
- Keep `internal/transcribe` as a leaf package with no upward dependencies.
- Use Go standard library for HTTP routing (no external router).
- Frontend uses Vue 3 Composition API with TypeScript.
- All user-facing strings should support internationalization (future goal).
