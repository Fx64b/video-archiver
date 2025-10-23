# Video Archiver

A self-hosted YouTube video archiver with a modern web interface. Download, manage, and organize your YouTube videos, playlists, and channels locally.

> [!CAUTION]
> This project is in early development (alpha stage). It may contain bugs, break unexpectedly, or have incomplete features. Use at your own risk.

## Overview

Video Archiver allows you to download and organize videos from YouTube for offline viewing. The application consists of a Go backend that handles the actual downloading and management of videos, plus a modern Next.js frontend that provides a clean interface for interacting with your media collection.

## Features

### Current Features
- Download videos, playlists, and channels from YouTube URLs
- Real-time download progress tracking via WebSocket
- Queue management for multiple downloads
- Modern responsive web interface
- Dark/Light mode support
- Dashboard with stats about your media collection
- Basic metadata extraction and display

### Planned Features
- Advanced download options (quality selection, format selection)
- Media conversion tools (video to audio, format conversion)
- Enhanced metadata management and organization
- Advanced search and categorization
- Mobile-friendly streaming interface
- User authentication and multiple user support
- Download scheduling
- API for integrations with other applications

## Technology Stack

- **Backend**: Go 1.23+ with Chi v5 router
- **Frontend**: Next.js 15+ with TypeScript
- **UI**: shadcn/ui components
- **State Management**: Zustand
- **Real-time Updates**: WebSocket
- **Database**: SQLite
- **Media Handling**: yt-dlp and ffmpeg
- **Security**: CORS protection, path traversal prevention

## Installation

### Prerequisites

- Docker and Docker Compose
- Git
- Bash shell (for running the management script)

### Quick Start

1. Clone the repository:
```bash
git clone https://github.com/Fx64b/video-archiver.git
cd video-archiver
```

2. Set environment variables:
```bash
cp web/.env.local.example web/.env.local
```

3. Start the application:
```bash
./run.sh
```

The web interface will be available at `http://localhost:3000`

### Run Script Options

```bash
./run.sh [options]

Options:
  --clear         Clear the database and downloads
  --build         Rebuild the containers and regenerate TypeScript types
  --debug         Enable debug logging in backend
  --backend-only  Start only the backend service
  --help|-h       Show help message
```

Common use cases:
- Fresh start: `./run.sh --clear --build`
- Development: `./run.sh --build --debug`
- Production: `./run.sh`

### Manual Setup (Without Docker)

#### Backend Requirements
- Go 1.23+
- SQLite
- yt-dlp
- ffmpeg

#### Frontend Requirements
- Node.js 20+
- pnpm

Refer to the backend and frontend directories for specific setup instructions.

## Project Structure

- `/services/backend`: Go backend service
  - Handles video downloads via yt-dlp
  - Manages download queue and progress tracking
  - Provides API endpoints and WebSocket updates
  - SQLite database for metadata storage
- `/web`: Next.js frontend application
  - Modern UI built with shadcn/ui
  - Real-time progress tracking
  - Download management interface
  - Statistics dashboard

## Configuration

The application can be configured through environment variables:

### Backend
- `DEBUG`: Enable debug logging (default: false)
- `DOWNLOAD_PATH`: Directory for downloaded media (default: ./data/downloads)
- `DATABASE_PATH`: Path to SQLite database (default: ./data/db/video-archiver.db)
- `PORT`: API server port (default: 8080)
- `WS_PORT`: WebSocket server port (default: 8081)
- `ALLOWED_ORIGINS`: Comma-separated list of allowed CORS origins (default: http://localhost:3000)

### Frontend
- `NEXT_PUBLIC_SERVER_URL`: URL for backend API (default: http://localhost:8080)
- `NEXT_PUBLIC_SERVER_URL_WS`: URL for WebSocket connection (default: ws://localhost:8081)

## Known Issues

### 403 Error When Downloading
After a certain number of downloads, YouTube may block requests that are not made through the official site or app.
This is a known issue: https://github.com/yt-dlp/yt-dlp/issues/11868

Possible workarounds:
- Use a VPN and switch servers when the limit is reached
- Wait a few hours before attempting downloads again
- Use a different IP address if possible

### Metadata Issues
Some metadata might not be properly extracted for certain types of content, particularly channels and playlists.

## Activity
![Alt](https://repobeats.axiom.co/api/embed/885b61cd30b55fb0b9635c6d9c46421f1bdbb262.svg "Repobeats analytics image")

## Contributing

We welcome contributions to Video Archiver! Here's how you can help:

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/AmazingFeature`)
3. Commit your changes using [Conventional Commits](https://www.conventionalcommits.org/) format
   - Examples: `feat(ui): Add video thumbnails`, `fix(api): Resolve playlist download issue`
4. Push to the branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

Please ensure your code follows the project's style and includes appropriate tests.

### Development Setup

For the best development experience:

1. Use `./run.sh --build --debug` to start the application with debug logging
2. For frontend-only changes, you can run the Next.js app directly:
   ```bash
   cd web
   pnpm install
   pnpm dev
   ```
3. For backend-only changes, you can use the `--backend-only` flag:
   ```bash
   ./run.sh --build --backend-only
   ```

## License

This project is licensed under the GNU General Public License v3.0 - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- [yt-dlp](https://github.com/yt-dlp/yt-dlp) for powerful video downloading capabilities
- [shadcn/ui](https://ui.shadcn.com/) for the beautiful UI components

## Roadmap

See the [open issues](https://github.com/Fx64b/video-archiver/issues) for a list of proposed features and known issues.

- Q3 2025: Media streaming interface & Tools
- Q4 2025: Advanced search and tagging system
