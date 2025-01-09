# Video Archiver

A self-hosted YouTube video archiver with a modern web interface. Download, manage, and organize your YouTube videos, playlists, and channels locally.

> [!CAUTION]
> This project is in early development (alpha stage). It may contain bugs, break unexpectedly, or have incomplete features. Use at your own risk.

## Features

### Current Features
- [x] Download videos from YouTube URLs
- [x] Real-time download progress tracking
- [x] Basic queue management
- [x] Modern web interface built with Next.js and shadcn/ui
- [x] Dark/Light mode support

### Planned Features
- Advanced download options (quality selection, format selection)
- Media conversion tools (video to audio, format conversion)
- Metadata extraction and management
- Search and categorization
- Mobile-friendly streaming interface

## Prerequisites

- Docker and Docker Compose
- Git
- Bash shell (for running the management script)

## Quick Start

1. Clone the repository:
```bash
git clone https://github.com/Fx64b/video-archiver.git
cd video-archiver
```

2. Start the application:
```bash
./run.sh
```

The web interface will be available at `http://localhost:3001`

## Run Script Options

The `run.sh` script provides several options to manage the application:

```bash
./run.sh [options]

Options:
  --clear    Clear the database and downloads
  --build    Rebuild the containers and regenerate TypeScript types
  --debug    Enable debug logging in backend
  --help|-h  Show help message
```

Common use cases:
- Fresh start: `./run.sh --clear --build`
- Development: `./run.sh --build --debug`
- Production: `./run.sh`

## Project Structure

- `/services/backend`: Go backend service
  - Handles video downloads
  - Manages download queue
  - Provides WebSocket updates
- `/web`: Next.js frontend application
  - Modern UI built with shadcn/ui
  - Real-time progress tracking
  - Download management interface

## Development

The project uses:
- Backend: Go 1.23 with Chi router
- Frontend: Next.js 15 with TypeScript
- UI: shadcn/ui components
- State Management: Zustand
- Real-time Updates: WebSocket
- Database: SQLite

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/AmazingFeature`)
3. Commit your changes (`git commit -m 'feat(scope): Add some AmazingFeature'`)
4. Push to the branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

## License

This project is licensed under the GNU General Public License v3.0 - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- [yt-dlp](https://github.com/yt-dlp/yt-dlp) for video downloading capabilities
- [shadcn/ui](https://ui.shadcn.com/) for the beautiful UI components