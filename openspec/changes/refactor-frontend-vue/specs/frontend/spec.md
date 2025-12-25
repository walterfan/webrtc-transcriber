## ADDED Requirements

### Requirement: Modern Frontend Stack
The frontend SHALL be built using Vue.js 3, TypeScript, and Tailwind CSS.

#### Scenario: Developer Setup
- **WHEN** a developer runs the build command
- **THEN** Vite compiles the Vue/TS source code
- **AND** produces optimized static assets in the output directory

### Requirement: User Authentication Interface
The system SHALL provide a secure login interface.

#### Scenario: Successful Login
- **WHEN** the user enters valid credentials
- **THEN** the system authenticates the user
- **AND** redirects to the main dashboard
- **AND** displays a "Welcome" message

### Requirement: WebRTC Recording & Transcription
The system SHALL support real-time audio recording and transcription control via the UI.

#### Scenario: Start Recording
- **WHEN** the user clicks "Start"
- **THEN** the browser captures microphone input
- **AND** establishes a WebRTC connection to the server
- **AND** displays a live audio waveform
- **AND** updates the recording timer

#### Scenario: Transcribe Only
- **WHEN** the user selects "Transcribe" (without Record)
- **AND** selects existing files
- **THEN** the system requests transcription for those files
- **AND** updates the results table with the text output

### Requirement: File Management
The system SHALL list available recordings and transcriptions.

#### Scenario: List Files
- **WHEN** the dashboard loads
- **THEN** it retrieves the list of files from the server
- **AND** groups audio and text files together
- **AND** displays them sorted by date (newest first)

#### Scenario: Audio Playback
- **WHEN** the user clicks play on an audio file
- **THEN** the custom audio player plays the file
- **AND** shows playback progress

