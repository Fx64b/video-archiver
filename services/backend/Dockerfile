FROM golang:1.24-alpine AS builder

# Install build dependencies
RUN apk add --no-cache gcc musl-dev

WORKDIR /build

# Copy go mod files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application (CGO required for mattn/go-sqlite3)
RUN CGO_ENABLED=1 go build -o main ./cmd/api

FROM alpine:latest

# Install runtime dependencies
RUN apk add --no-cache ffmpeg python3 curl \
    && curl -L https://github.com/yt-dlp/yt-dlp/releases/download/2025.03.31/yt-dlp -o /usr/local/bin/yt-dlp \
    && chmod +x /usr/local/bin/yt-dlp

# Create non-root user and setup directories
RUN addgroup -S appgroup && adduser -S appuser -G appgroup
WORKDIR /app

# Copy files from builder
COPY --from=builder /build/main .
COPY --from=builder /build/db/schema.sql ./db/

# Setup permissions
RUN mkdir -p /app/data/downloads /app/data/db /app/cache && \
    chown -R appuser:appgroup /app && \
    chmod -R 755 /app/data && \
    chmod -R 777 /app/data/db

# Set environment and switch user
ENV XDG_CACHE_HOME=/app/cache
USER appuser

EXPOSE 8080 8081

CMD ["./main"]