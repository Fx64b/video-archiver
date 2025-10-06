# Backend Testing Guide

This document provides information about the testing infrastructure for the video-archiver backend.

## Overview

The backend has comprehensive unit tests covering:

- **Domain layer**: Business logic and data models
- **Repository layer**: Database operations (SQLite)
- **Service layer**: Download service, metadata extraction, progress tracking
- **API layer**: HTTP handlers and request/response handling
- **Utilities**: Version checker, statistics calculator


## Running Tests

### Using Make (Recommended)

```bash
# Run all tests
make test

# Run tests with verbose output
make test-verbose

# Run tests with coverage report
make test-coverage

# Run specific package tests
make test-domain
make test-repository
make test-services
make test-handlers
make test-utils

# Clean test cache
make clean

# List all tests
make test-list

# Run tests in parallel (faster)
make test-parallel
```

### Using the run.sh Script

```bash
# From project root
./run.sh --test              # Run all tests
./run.sh --test-verbose      # Verbose output
./run.sh --test-coverage     # Generate coverage report
```

### Using Go Directly

```bash
cd services/backend

# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run specific package
go test ./internal/domain

# Run tests with race detection
go test -race ./...

# Run tests with verbose output
go test -v ./...
```

## Test Structure

### Test Helpers

Located in `internal/testutil/`:

- `helpers.go`: Test utilities, mock repository, test database setup
- `fixtures.go`: Sample yt-dlp output and metadata JSON

### Test Files

Each package has corresponding `*_test.go` files:

```
internal/
├── domain/
│   └── job_test.go
├── repositories/sqlite/
│   └── job_repository_test.go
├── services/
│   ├── download/
│   │   └── progress_test.go
│   └── metadata/
│       └── metadata_test.go
├── api/handlers/
│   └── handlers_test.go
└── util/
    ├── statistics/
    │   └── statistics_test.go
    └── version/
        └── checker_test.go
```

## Writing Tests

### Example Test Structure

```go
func TestFunctionName(t *testing.T) {
    // Setup
    mockRepo := testutil.NewMockJobRepository()

    // Test data
    job := testutil.CreateTestJob("test-id", "https://youtube.com/watch?v=test")

    // Execute
    err := mockRepo.Create(job)

    // Assert
    if err != nil {
        t.Fatalf("Create() error = %v", err)
    }
}
```

### Using Test Helpers

```go
// Create in-memory test database
db := testutil.CreateTestDB(t)
defer db.Close()

// Create mock repository
mockRepo := testutil.NewMockJobRepository()

// Create test data
job := testutil.CreateTestJob("id", "url")
videoMeta := testutil.CreateTestVideoMetadata()
playlistMeta := testutil.CreateTestPlaylistMetadata()
channelMeta := testutil.CreateTestChannelMetadata()
```

### Testing with yt-dlp Output

Use fixtures from `testutil/fixtures.go`:

```go
// Use sample yt-dlp output
output := testutil.YtDlpProgressOutput

// Use sample metadata JSON
metadataJSON := testutil.VideoMetadataJSON
```

## Mocking

### Mock Repository

The `MockJobRepository` provides in-memory implementations of all repository methods:

```go
mockRepo := testutil.NewMockJobRepository()

// Use like a real repository
job := &domain.Job{ID: "test", URL: "https://youtube.com/watch?v=test"}
mockRepo.Create(job)
mockRepo.Update(job)
retrieved, _ := mockRepo.GetByID("test")
```

### Database Tests

SQLite repository tests use an in-memory database:

```go
db := testutil.CreateTestDB(t)  // Creates :memory: SQLite DB with schema
defer db.Close()

repo := sqlite.NewJobRepository(db)
// Test repository methods...
```

## Test Fixtures

### Mock yt-dlp Output

Available in `testutil/fixtures.go`:

- `VideoMetadataJSON`: Sample video metadata
- `PlaylistMetadataJSON`: Sample playlist metadata
- `ChannelMetadataJSON`: Sample channel metadata
- `YtDlpProgressOutput`: Sample progress output
- `YtDlpPlaylistProgressOutput`: Playlist download progress
- `YtDlpAlreadyDownloadedOutput`: Already downloaded scenario
- `ArchiveFileContent`: Sample archive file content

## Coverage Report

Generate an HTML coverage report:

```bash
make test-coverage
# Opens coverage.html in browser
```

Or manually:

```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

## Continuous Integration

Tests are automatically run in GitHub Actions CI pipeline on:

- **Pull requests** to `main` and `develop` branches
- **Pushes** to `main` and `develop` branches
- **Before Docker builds** - Docker Compose build only runs after tests pass

### CI Workflows

1. **Test Workflow** (`.github/workflows/test.yml`):
   - Runs all backend unit tests with race detection
   - Generates coverage reports
   - Uploads coverage artifacts
   - Displays coverage summary in GitHub Actions UI

2. **Build and Lint Workflow** (`.github/workflows/build-lint.yml`):
   - Runs backend tests first
   - Only builds Docker images if tests pass
   - Validates Docker Compose configuration

### Viewing CI Results

- Test results appear in the GitHub Actions tab
- Coverage summary is displayed in the workflow summary
- Coverage artifacts are available for download for 30 days
- Failed tests block the Docker build process

### Local Pre-commit Testing

Before pushing code, run tests locally:

```bash
# Quick test
./run.sh --test

# With coverage
./run.sh --test-coverage

# Or using Make
cd services/backend
make test
```

## Best Practices

1. **Isolation**: Each test should be independent
2. **Cleanup**: Use `t.TempDir()` for temporary files
3. **Table-driven**: Use table-driven tests for multiple scenarios
4. **Descriptive names**: Test names should clearly describe what they test
5. **Fast tests**: Keep unit tests fast (< 1 second each)
6. **Mock external dependencies**: Mock yt-dlp, network calls, etc.

## Troubleshooting

### Tests timing out

If tests hang or timeout:

```bash
# Run with timeout
go test -timeout=10s ./...

# Run specific package with verbose output
go test -v -timeout=10s ./internal/domain
```

### Cache issues

Clear the test cache:

```bash
make clean
# or
go clean -testcache
```

### Database lock errors

Ensure tests properly clean up database connections:

```go
db := testutil.CreateTestDB(t)
defer db.Close()  // Always defer Close()
```

## Future Improvements

- [ ] Integration tests with real yt-dlp
- [ ] End-to-end API tests
- [ ] Performance/benchmark tests
- [ ] Increase coverage to 80%+
- [ ] Add mutation testing
- [ ] Docker-based test environment

## Resources

- [Go Testing Package](https://pkg.go.dev/testing)
- [Table Driven Tests](https://github.com/golang/go/wiki/TableDrivenTests)
- [Testify Assertions](https://github.com/stretchr/testify)
