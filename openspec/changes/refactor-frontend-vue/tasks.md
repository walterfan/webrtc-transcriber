## 1. Setup & Config
- [x] 1.1 Initialize Vue+TS+Vite project in `frontend/` directory.
- [x] 1.2 Install and configure Tailwind CSS.
- [x] 1.3 Configure Vite to build to `frontend/dist`.
- [x] 1.4 Update `Makefile` to include frontend build commands (`npm install`, `npm run build`).

## 2. Core Logic (Composables)
- [x] 2.1 Implement `useAuth` (login, logout, status check).
- [x] 2.2 Implement `useWebRTC` (peer connection, data channel, signaling).
- [x] 2.3 Implement `useAudioVisualization` (waveform rendering).
- [x] 2.4 Implement `useFileManager` (fetch files, delete, transcribe).

## 3. UI Components
- [x] 3.1 Create `LoginForm.vue`.
- [x] 3.2 Create `Navbar.vue` and `Footer.vue` with Tailwind styling.
- [x] 3.3 Create `AudioPlayer.vue` (custom player logic).
- [x] 3.4 Create `FileTable.vue` and `FileRow.vue`.
- [x] 3.5 Create `Waveform.vue`.
- [x] 3.6 Create main `Dashboard.vue` layout.

## 4. Integration
- [x] 4.1 Assemble `App.vue` with routing (Login vs Dashboard).
- [x] 4.2 Update `cmd/transcribe-server/main.go` to serve `frontend/dist`.
- [x] 4.3 Verify full flow: Login -> Record -> Transcribe -> Playback.
- [x] 4.4 Clean up old `web/js` and `web/vendor` files.
