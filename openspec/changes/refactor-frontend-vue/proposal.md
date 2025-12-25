# Change: Refactor Frontend to Vue.js + TypeScript

## Why
The current frontend is implemented in a single `app.js` file using React without JSX. This makes maintenance difficult, lacks type safety, and misses out on modern build tool optimizations. A move to Vue.js 3 with TypeScript and Vite will improve developer experience, code organization, and performance.

## What Changes
- **Tech Stack**: Replace React (no-JSX) + Bulma with Vue.js 3 + TypeScript + Tailwind CSS + Vite.
- **Directory Structure**: Create a new `frontend/` directory for the source code.
- **Build Process**: Introduce a build step using Vite, outputting to `frontend/dist`.
- **Server Configuration**: Update the Go server to serve static files from `frontend/dist`.
- **Code Organization**: Split the monolithic `app.js` into modular Vue components and composables.

## Impact
- **Affected Specs**: `frontend` (New capability)
- **Affected Code**:
    - `web/`: Existing files will be replaced/moved.
    - `cmd/transcribe-server/main.go`: Static file serving path update.
    - `Makefile`: Add frontend build instructions.

