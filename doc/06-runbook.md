# Operations Runbook

## One-Command Startup

```bash
# Build frontend + backend and run
make && ./webrtc-transcriber
```

The server starts on <http://localhost:9070> by default.

## Prerequisites

### System Dependencies

| Dependency | Required | Purpose |
|------------|----------|---------|
| Go 1.12+ | Yes | Build the backend |
| Node.js 18+ | Yes | Build the frontend |
| Python 3.8+ | For Whisper | Run whisper-ctranslate2 |
| libopus-dev | Yes | Opus audio codec (C library) |
| pkg-config | Yes | Locate libopus headers |

### Installing Dependencies

::::{tab-set}

:::{tab-item} macOS

```bash
brew install opus pkg-config go node python3
pip install whisper-ctranslate2
```
:::

:::{tab-item} Ubuntu/Debian

```bash
sudo apt install libopus-dev pkg-config golang-go nodejs npm python3 python3-pip
pip install whisper-ctranslate2
```
:::

::::

## Build Commands

| Command | Description |
|---------|-------------|
| `make` | Build frontend (npm install + build) and backend (go build) |
| `make build-frontend` | Build only the Vue frontend |
| `make build-backend` | Build only the Go binary |
| `go run ./cmd/transcribe-server/main.go` | Run without building a binary |

## Development Mode

### Backend

```bash
go run ./cmd/transcribe-server/main.go --http.port=9070
```

### Frontend (with hot reload)

```bash
cd frontend
npm run dev
```

Vite proxies API requests to `localhost:9070` (configured in `vite.config.ts`).

## Configuration

### Minimal Setup

```bash
# Create .env with at least one account
echo 'accounts=admin:changeme' > .env

# Build and run
make && ./webrtc-transcriber
```

### Full Configuration

```bash
cp env.example .env
# Edit .env with your values
```

### CLI Flags

```bash
./webrtc-transcriber \
  --vendor=whisper \
  --model=small \
  --language=auto \
  --output=./recordings \
  --http.port=9070 \
  --stun.server=stun:stun.l.google.com:19302 \
  --keep_wav \
  --keep_txt
```

## Common Troubleshooting

### Opus build errors

```
# error: opus.h: No such file or directory
```

**Fix**: Install libopus development headers.

```bash
# macOS
brew install opus

# Ubuntu
sudo apt install libopus-dev
```

### Whisper not found

```
Whisper service not available: exec: "whisper-ctranslate2": executable file not found in $PATH
Falling back to Recorder service
```

**Fix**: Install whisper-ctranslate2 or set `WHISPER_PATH` in `.env`.

```bash
pip install whisper-ctranslate2
```

### Port already in use

```
listen tcp :9070: bind: address already in use
```

**Fix**: Use a different port or kill the existing process.

```bash
./webrtc-transcriber --http.port=9071
# or
lsof -ti :9070 | xargs kill
```

### No accounts configured

```
Warning: No accounts configured in .env file
```

**Fix**: Add accounts to `.env`.

```bash
echo 'accounts=admin:password' >> .env
```

### WebRTC connection fails behind NAT

**Fix**: Ensure the STUN server is reachable. The default (`stun.l.google.com:19302`) requires internet access.

```bash
./webrtc-transcriber --stun.server=stun:your-stun-server:3478
```

### Frontend shows blank page

**Fix**: Ensure the frontend is built. The Go server serves files from `frontend/dist/`.

```bash
make build-frontend
```

## Monitoring

No built-in metrics or health-check endpoint. Monitor the process with standard OS tools:

```bash
# Check if running
pgrep webrtc-transcriber

# Check disk usage for recordings
du -sh recordings/

# Tail logs
./webrtc-transcriber 2>&1 | tee server.log
```
