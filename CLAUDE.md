# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Development Commands

### Docker-based Development (Recommended)
- `./run.sh` - Start the full application
- `./run.sh --build` - Rebuild containers and regenerate TypeScript types
- `./run.sh --debug` - Start with debug logging enabled
- `./run.sh --clear --build` - Fresh start (clears database and downloads)
- `./run.sh --backend-only` - Start only the backend service

### Frontend Development
- `cd web && pnpm install` - Install dependencies
- `cd web && pnpm dev` - Start Next.js development server
- `cd web && pnpm build` - Build the frontend for production
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
- **Frontend**: Next.js 15 app at `web/` with TypeScript, shadcn/ui components
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
- **Next.js App Router** with TypeScript
- **Dashboard layout**: Sidebar navigation with multiple pages
- **Component structure**: shadcn/ui components in `web/components/ui/`
- **Pages**: Dashboard, Downloads, Channels, Settings, Tools
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
- **Frontend**: Next.js 15, TypeScript, shadcn/ui, Zustand, WebSocket
- **Development**: Docker Compose for orchestration

### Progress Tracking Implementation
- **Template-based progress parsing**: Uses structured yt-dlp output templates
- **Video state tracking**: Individual VideoDownloadState for each video ID
- **Phase detection**: Distinguishes between video, audio, and merging phases
- **Accurate progress calculation**: Video (80%) + Audio (20%) for single videos

### Environment Configuration
- Backend: `DEBUG`, `DOWNLOAD_PATH`, `DATABASE_PATH`, `PORT`, `WS_PORT`
- Frontend: `NEXT_PUBLIC_SERVER_URL`, `NEXT_PUBLIC_SERVER_URL_WS`
- Default development URLs: Backend on :8080, WebSocket on :8081, Frontend on :3000

## Future Improvements

### WebSocket Reconnection
- **Feature**: When the UI reconnects (WebSocket reconnection or page refresh), recent downloads should be reloaded
- **Context**: Currently when users refresh the page or reconnect after network issues, they lose visibility into recent download progress and status
- **Implementation**: Add logic to fetch and display recent downloads on WebSocket connection establishment
- to memorize in the future somewhen i want to add / adjust metdata for this app, for exmaple the tab title