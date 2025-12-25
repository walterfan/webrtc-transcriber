## ADDED Requirements

### Requirement: Video Recording
The system SHALL support receiving and recording video streams from the WebRTC client.

#### Scenario: Receive Video
- **WHEN** a client offers a video track (VP8 or H.264)
- **THEN** the server accepts the video track
- **AND** writes the incoming RTP packets to a local video file (e.g., .ivf or .h264)

### Requirement: MP4 Output
The system SHALL produce a single MP4 file containing the recorded audio and video sessions.

#### Scenario: Generate MP4
- **WHEN** the WebRTC session ends
- **THEN** the system executes a muxing process (using ffmpeg)
- **AND** combines the recorded audio and video files into a single `.mp4` file
- **AND** ensures audio and video are synchronized

