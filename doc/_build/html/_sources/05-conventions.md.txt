# Engineering Conventions

## Code Style

### Go

- Follow the standard Go formatting enforced by `gofmt`.
- Packages under `internal/` are not importable by external code.
- Interfaces defined in the same package as their primary consumer.
- Error messages start with lowercase and do not end with punctuation (Go convention).
- Use `context.Context` for cancellation propagation in service constructors.

### TypeScript / Vue

- Vue 3 Composition API with `<script setup lang="ts">`.
- Composables in `frontend/src/composables/` with `use` prefix.
- Tailwind CSS utility classes for styling (no separate CSS files per component).
- Vite as the build tool; proxy configuration in `vite.config.ts` for local development.

## Error Handling

### Go Backend

- Functions return `(value, error)` pairs.
- `fmt.Errorf("context: %w", err)` for error wrapping.
- Fatal errors at startup use `log.Fatalf`.
- HTTP handlers return status codes without custom error types.
- Whisper unavailability triggers graceful fallback to Recorder.

### Frontend

- Composables handle errors internally and expose reactive error state.
- Network errors displayed to the user via UI notifications.
- WebRTC connection failures trigger reconnection logic.

## Logging

- **Backend**: Go standard `log` package. No structured logging, no log levels.
- **Key events logged**: Account loading, vendor selection, session creation/destruction, file operations, errors.
- **No trace IDs or correlation IDs** in the current implementation.

## Configuration Management

| Source | Priority | Examples |
|--------|----------|---------|
| CLI flags | Highest | `--vendor`, `--model`, `--http.port` |
| `.env` file | Medium | `AZURE_SPEECH_KEY`, `accounts` |
| Defaults in code | Lowest | Port `9070`, model `small`, vendor `whisper` |

- Secrets (API keys, passwords) go in `.env` (never committed).
- `env.example` serves as a template.
- No YAML/TOML/JSON config files.

## Feature Flags

No feature flag system. Operation modes (Full, Record Only, Transcribe Only) are selected per-session via the frontend UI.

## Versioning

- No version tagging or release process currently in place.
- Go module version: `github.com/walterfan/webrtc-transcriber` (no version suffix).

## Anti-Patterns

:::{admonition} Avoid
:class: warning

- **Inline handlers in main.go**: The `/files` and `/delete/` handlers are defined inline as closures. Prefer extracting them into the `session` package or a new `files` package for testability.
- **Plaintext passwords**: Accounts are stored as plaintext in the `accounts` env var. Use hashed passwords for any non-prototype deployment.
- **No input validation on SDP**: The session handler trusts the SDP offer without size or format validation.
- **Global mutable state**: `sessionStore` and `accounts` are package-level variables. Consider dependency injection for testability.
:::

## Git Workflow

- Main branch: `main`.
- Change proposals documented in `openspec/changes/`.
- No CI/CD pipeline currently configured.
