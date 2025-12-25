# Design: Frontend Refactor (Vue + TS + Tailwind)

## Context
The current frontend is a ~1600 line JavaScript file using `React.createElement`. This was likely done to avoid a build step, but it scales poorly. We want to modernize this with a standard Vue 3 toolchain.

## Goals / Non-Goals
- **Goal**: Replicate existing functionality 1:1 (Recording, Transcription, File Management, Auth).
- **Goal**: establish a scalable component architecture.
- **Goal**: Type safety with TypeScript.
- **Non-Goal**: Redesigning the UX flow (visual refresh with Tailwind is okay, but logic remains same).

## Decisions

### Decision: Vite + Vue 3 + TypeScript
We will use the standard Vite template for Vue TS.
- **Why**: Fast dev server, optimized production build, industry standard for Vue.

### Decision: Tailwind CSS
Replacing Bulma with Tailwind CSS.
- **Why**: Requested by user. Provides utility-first styling which fits well with component-based architecture.
- **Implication**: We need to rewrite all styles. Bulma's `.button`, `.navbar`, etc., will be replaced with Tailwind utilities (e.g., `px-4 py-2 bg-blue-500 text-white rounded`).

### Decision: State Management
We will use Vue's Composition API (`ref`, `reactive`) for local state and simple shared state.
- **Why**: The app's state is relatively flat (Auth, Current Session, File List). A full store (Pinia) might be overkill but we can introduce it if state complexity grows. For now, simple composables (e.g., `useAuth`, `useRecorder`) are sufficient.

### Decision: Directory Structure
```
frontend/
├── src/
│   ├── components/    # UI Components (Button, Navbar, etc.)
│   ├── composables/   # Logic (useWebRTC, useAudioPlayer)
│   ├── assets/        # Images, Fonts
│   ├── App.vue
│   └── main.ts
├── index.html
├── vite.config.ts
└── tailwind.config.js
```
The build output will go to `frontend/dist/`. The Go server will serve `frontend/dist/` as the static root.

## Migration Plan
1.  Initialize `frontend` project.
2.  Implement `useAuth` and `LoginForm`.
3.  Implement `useWebRTC` porting logic from `app.js`.
4.  Build main UI shell (Navbar, Footer) with Tailwind.
5.  Implement `AudioPlayer` and `FileTable`.
6.  Switch Go server to serve new assets.
7.  Verify feature parity.
8.  Delete old `web/js` and `web/vendor`.

## Open Questions
- **Icons**: Currently using FontAwesome. We can continue using it (via npm package) or switch to a Vue icon library (e.g., `heroicons` or `lucide-vue-next`). *Decision: Use `lucide-vue-next` for better Tailwind integration, or keep FontAwesome if specific icons are needed.* (Stick to FontAwesome for now to minimize visual diff, or switch if easier). -> Let's switch to **Lucide** or **Heroicons** for a more modern feel matching Tailwind, unless specific FA icons are critical.

## Risks / Trade-offs
- **Risk**: Regression in WebRTC handling during porting.
    - *Mitigation*: Careful testing of the `useWebRTC` composable against the existing backend.

