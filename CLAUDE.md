# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Development Commands

### Docker-based Development (Recommended)
- `./run.sh dev` - Hot-reload dev stack: backend via Air (<1s rebuilds), frontend via Vite HMR, both bind-mounted
- `./run.sh` - Start the full application (builds images on first run)
- `./run.sh build` - Rebuild images and start (use after pulling changes)
- `./run.sh stop` - Stop the application
- `./run.sh logs [service]` - Follow container logs
- `./run.sh test [coverage|verbose]` - Run backend tests
- `./run.sh types` - Regenerate `web/types/index.ts` from Go domain structs (commit the result)
- `./run.sh reset` - Stop and delete all data (database, downloads)
- `./run.sh --debug` / `--backend-only` / `--detach` - Options for start/build
- Dockerfiles use BuildKit cache mounts (Go build cache, pnpm store), so rebuilds are incremental

### Frontend Development
- `cd web && pnpm install` - Install dependencies
- `cd web && pnpm dev` - Start Vite development server
- `cd web && pnpm build` - Typecheck and build the frontend for production
- `cd web && pnpm preview` - Serve the production build locally
- `cd web && pnpm test` - Run Vitest test suite
- `cd web && pnpm lint` - Run ESLint
- `cd web && pnpm format` - Format code with Prettier

### Backend Development
- Backend is Go-based, located in `services/backend/`
- Uses `tygo generate` to generate TypeScript types from Go structs
- Types are generated to `web/types/index.ts` during build process

## Architecture Overview

### Project Structure
- **Monorepo** with separate backend and frontend services
- **Backend**: Go service at `services/backend/` with Chi router, SQLite database
- **Frontend**: Vite + React SPA at `web/` with TypeScript, shadcn/ui components
- **Real-time updates**: WebSocket connection between backend and frontend
- **State management**: Zustand store at `web/store/appState.ts`

### Backend Architecture
- **Domain-driven design** with clear separation of concerns
- **Key domains**: Jobs (downloads), Metadata (video/playlist/channel info)
- **Repository pattern**: SQLite repository implementations
- **Services**: Download service handles yt-dlp integration and job queue
- **WebSocket hub**: Real-time progress updates to frontend
- **API endpoints**: REST API for job management and metadata retrieval

### Frontend Architecture
- **Vite + React SPA** with TypeScript and React Router (routes in `web/src/App.tsx`)
- **Entry points**: `web/index.html`, `web/src/main.tsx`; pages in `web/src/pages/`
- **Dashboard layout**: Sidebar navigation with multiple pages
- **Component structure**: shadcn/ui components in `web/components/ui/`
- **Pages**: Overview, Dashboard, Downloads, Settings, Tools
- **Real-time updates**: WebSocket integration for download progress
- **State management**: Zustand for global app state

### Data Flow
1. User submits URL via frontend
2. Backend creates Job entity, queues download with yt-dlp
3. Download service processes job, extracts metadata
4. Progress updates sent via WebSocket to frontend
5. Completed metadata stored in SQLite, available via API

### Key Types and Entities
- **Job**: Download job with status, progress, timestamps
- **Metadata**: Interface implemented by VideoMetadata, PlaylistMetadata, ChannelMetadata
- **JobWithMetadata**: Combines job info with extracted content metadata
- **ProgressUpdate**: Real-time download progress information

### Technology Stack
- **Backend**: Go 1.23+ with Chi router, SQLite, yt-dlp, ffmpeg
- **Frontend**: Vite 6, React 19, TypeScript, shadcn/ui, Zustand, React Router, WebSocket
- **Development**: Docker Compose for orchestration

### Progress Tracking Implementation
- **Template-based progress parsing**: Uses structured yt-dlp output templates
- **Video state tracking**: Individual VideoDownloadState for each video ID
- **Phase detection**: Distinguishes between video, audio, and merging phases
- **Accurate progress calculation**: Video (80%) + Audio (20%) for single videos

### Environment Configuration
- Backend: `DEBUG`, `DOWNLOAD_PATH`, `PROCESSED_PATH`, `DATABASE_PATH`, `PORT` (REST and the /ws WebSocket share the port)
- Frontend: no config needed — same-origin `/api` URLs, proxied by Vite in dev and nginx in prod; `VITE_SERVER_URL`/`VITE_SERVER_URL_WS` are optional build-time overrides (see `web/lib/env.ts`)
- Default development URLs: Backend on :8080 (REST + /ws), Frontend on :3000
- The SQLite driver is pure Go (modernc.org/sqlite); build with `CGO_ENABLED=0`. Do not reintroduce cgo dependencies — they cost minutes of cold build time

## Future Improvements

No open items currently tracked.