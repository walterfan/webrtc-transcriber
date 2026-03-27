# Project Overview

## Purpose

WebRTC Transcriber is a real-time speech-to-text application that captures audio from a browser via WebRTC, streams it to a Go backend, and feeds it into an AI transcription engine (primarily OpenAI Whisper running locally). Transcription results are pushed back to the browser over a WebRTC DataChannel with low latency.

The project originated as a fork of [webrtc-speech-to-text](https://github.com/rviscarra/webrtc-speech-to-text) with significant additions: a Vue 3 frontend, multi-vendor transcription support, audio recording/playback, and file management.

## Business Boundaries

### What We Do

- Capture microphone audio in the browser and stream it to the server via WebRTC (Opus codec).
- Decode Opus to PCM on the server and pipe it to a transcription backend.
- Return transcription results to the browser in real time over a DataChannel.
- Record audio to WAV files on disk for later playback or re-transcription.
- Provide a web UI for controlling recording, transcription, language selection, and file management.
- Support multiple transcription vendors: Whisper (default, local), Google Speech, Azure Speech, Baidu, and Xunfei.

### What We Don't Do

- No user registration or persistent user database -- accounts are configured via environment variables.
- No multi-tenant isolation -- all authenticated users share the same recordings directory.
- No horizontal scaling or clustering -- single-process, single-node deployment.
- No video capture or screen sharing.
- No real-time collaboration or shared transcription sessions between users.

## Key User Roles

- **End User**: Opens the web UI, logs in, records audio, and views transcription results.
- **Administrator**: Configures accounts, selects the transcription vendor, and manages the server process.

## Core Use Cases

1. **Full Mode** (default): Record audio and transcribe simultaneously. Results appear in real time.
2. **Record Only**: Capture audio to a WAV file without transcription.
3. **Transcribe Only**: Transcribe a previously recorded WAV file (frontend-initiated; backend endpoint pending).
4. **File Management**: List, play, download, and delete recordings from the web UI.

## Technology Stack

| Layer | Technology | Version |
|-------|-----------|---------|
| Language (backend) | Go | 1.12+ |
| WebRTC | Pion WebRTC | v2.0.15 |
| Audio codec | Opus (hraban/opus) | v2 |
| HTTP routing | Go `net/http` (stdlib) | -- |
| Config loading | godotenv | v1.5.1 |
| Frontend framework | Vue 3 + TypeScript | 3.x |
| Build tool | Vite | 7.x |
| CSS framework | Tailwind CSS | 4.x |
| Transcription (default) | whisper-ctranslate2 (Python) | latest |
| Cloud alternatives | Google Speech, Azure Speech, Baidu, Xunfei | various |

## Deployment Model

**Single-process monolith.** The Go binary serves both the API and the static frontend assets (`frontend/dist`). There is no container orchestration, no database, and no message queue. The server listens on a single HTTP port (default `9070`).

## Quality Targets

| Dimension | Target |
|-----------|--------|
| Latency | Transcription results visible within 2-5 seconds of speech (Whisper, `small` model) |
| Availability | Proof-of-concept; no HA requirements |
| Accuracy | Depends on Whisper model size; `small` is the default trade-off |
| Privacy | All processing local by default (Whisper); no data leaves the machine |
| Browser support | Chrome 75+, Firefox 67+, Safari 12.1+ |
