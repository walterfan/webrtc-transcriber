## 1. Implementation
- [ ] 1.1 Add `ffmpeg` check in server startup (warn if missing).
- [ ] 1.2 Update `PionRtcService` to include Video Transceiver (RecvOnly) in `CreatePeerConnection`.
- [ ] 1.3 Implement `handleVideoTrack` goroutine to write to IVF (VP8) or H264 file.
- [ ] 1.4 Update `handleAudioTrack` to optionally write to OGG file (in addition to existing WAV/transcription pipe) for easier muxing, OR reuse existing WAV.
- [ ] 1.5 Implement `stopRecording` logic to trigger `ffmpeg` muxing command.
- [ ] 1.6 Verify MP4 output with H.264 and VP8 sources.

