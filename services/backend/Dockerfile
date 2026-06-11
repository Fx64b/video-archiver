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

# Deno is required by recent yt-dlp versions to solve YouTube's JavaScript
# challenges (nsig / PO tokens). The official Deno binary is built against
# glibc, so gcompat + libstdc++ let it run on Alpine's musl libc.
COPY --from=denoland/deno:bin-2.1.4 /deno /usr/local/bin/deno

# Install runtime dependencies
RUN apk add --no-cache ffmpeg python3 curl gcompat libstdc++ \
    && curl -L https://github.com/yt-dlp/yt-dlp/releases/download/2026.03.17/yt-dlp -o /usr/local/bin/yt-dlp \
    && chmod +x /usr/local/bin/yt-dlp /usr/local/bin/deno

# Create non-root user and setup directories
RUN addgroup -S appgroup && adduser -S appuser -G appgroup
WORKDIR /app

# Copy files from builder
COPY --from=builder /build/main .
COPY --from=builder /build/db/schema.sql ./db/

# Setup permissions
RUN mkdir -p /app/data/downloads /app/data/db /app/data/processed /app/cache && \
    chown -R appuser:appgroup /app && \
    chmod -R 755 /app/data && \
    chmod -R 777 /app/data/db /app/data/processed

# Set environment and switch user
ENV XDG_CACHE_HOME=/app/cache
USER appuser

EXPOSE 8080 8081

CMD ["./main"]
