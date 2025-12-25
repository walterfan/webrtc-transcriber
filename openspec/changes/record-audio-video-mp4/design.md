# Design: Audio/Video Recording to MP4

## Context
The current system captures audio via WebRTC, decodes Opus to PCM, and saves it as WAV (and optionally transcribes it). Users want video support. WebRTC video typically comes as VP8 or H.264. MP4 is the desired container. Pion WebRTC can write to IVF (for VP8) or raw H.264 streams, but doesn't natively support MP4 muxing with audio/video synchronization out of the box without complex logic.

## Goals / Non-Goals
- **Goal**: Produce a synchronized MP4 file containing both audio and video from the session.
- **Goal**: Maintain existing real-time transcription capability (audio split).
- **Non-Goal**: Real-time streaming to RTMP/HLS (recording to file only).
- **Non-Goal**: GPU acceleration for encoding (rely on CPU/ffmpeg default).

## Decisions

### Decision: Post-Processing Muxing with FFmpeg
We will save the raw streams during the session and mux them after the session ends.
- **Why**:
    - **Simplicity**: Pion has built-in writers for IVF (VP8) and OGG (Opus).
    - **Robustness**: Decouples network I/O from complex container muxing. If the muxer crashes, we still have the raw dumps.
    - **Standard**: `ffmpeg` is the industry standard for this. Writing a compliant MP4 muxer in Go is significant effort.
- **Flow**:
    1.  **Session Start**: Open `session.ivf` (video) and `session.ogg` (audio) files.
    2.  **During Session**: Write RTP packets to these files using Pion's helpers (`ivfwriter`, `oggwriter`).
    3.  **Session End**: Close files. Execute `ffmpeg -i session.ivf -i session.ogg -c:v copy -c:a aac session.mp4`.
    4.  **Cleanup**: Delete raw files (optional/configurable).

### Decision: Codec Selection
- **Video**: Negotiate **VP8** as primary. It's the most compatible default for WebRTC.
- **Audio**: **Opus** (existing).
- **Container**: MP4. FFmpeg will transcode Opus to AAC (more standard for MP4) or keep Opus if MP4 support allows (modern players support Opus in MP4, but AAC is safer). Let's transcode audio to AAC for maximum compatibility. Video (VP8) can be put in MP4, but H.264 is native. Converting VP8 to H.264 is CPU intensive. We will try `copy` first (VP8 in MP4 is supported by many browsers/players, or we use MKV/WebM if MP4 is strictly requested but VP8 is used).
    - *Refinement*: If the user *strictly* needs MP4, we should prefer negotiating **H.264** if the browser supports it, to avoid transcoding. If H.264 is not available, we fall back to VP8 and transcode (slow) or wrap in WebM (fast).
    - *Plan*: Prioritize H.264 in SDP preferences. If we get H.264, save to `.h264`, then mux to MP4 (fast). If VP8, save to `.ivf`, transcode to H.264 MP4 (slow) OR just save as WebM (fast). For this proposal "to MP4", we will assume transcoding if necessary, or H.264 negotiation.

## Risks / Trade-offs
- **Risk**: FFmpeg dependency.
    - *Mitigation*: Check for `ffmpeg` binary on startup. Fail gracefully or disable recording if missing.
- **Risk**: Transcoding CPU usage.
    - *Mitigation*: Prefer H.264 negotiation.

## Migration Plan
- N/A (New feature)

