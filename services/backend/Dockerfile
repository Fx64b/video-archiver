FROM golang:1.24-alpine AS builder

# Install build dependencies
RUN apk add --no-cache gcc musl-dev

WORKDIR /build
COPY . .
RUN go build -o main ./cmd/api

FROM alpine:latest

# Install runtime dependencies
RUN apk add --no-cache ffmpeg python3 curl \
    && curl -L https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp -o /usr/local/bin/yt-dlp \
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