# Change: Record Audio and Video to MP4

## Why
Users need to record both audio and video sessions for playback and archival. Currently, the system only supports audio recording (saved as WAV/TXT). Outputting to MP4 provides a widely compatible format for sharing and playback.

## What Changes
- Add video support to WebRTC negotiation (SDP).
- Implement video track handling in the Pion RTC service.
- Record video stream to a temporary file (e.g., IVF for VP8 or raw H.264).
- Mux the recorded audio and video into a single MP4 file after the session ends.
- Dependency: Requires `ffmpeg` installed on the server for muxing.

## Impact
- **Affected Specs**: `recording` (New capability)
- **Affected Code**:
    - `internal/rtc/pion.go`: Add video transceiver and track handling.
    - `internal/rtc/service.go`: Update interfaces if needed.
    - `cmd/transcribe-server/main.go`: Ensure ffmpeg is available or handle errors.

