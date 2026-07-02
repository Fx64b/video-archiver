# Build stage: compile the Go binary.
# gcc and musl-dev are needed because mattn/go-sqlite3 compiles SQLite via cgo.
FROM golang:1.24-alpine AS builder

RUN apk add --no-cache gcc musl-dev

WORKDIR /build

# Resolve modules first so they are cached independently of source changes.
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

COPY . .

# The cache mounts persist the module cache and compiled objects (including
# the expensive cgo build of SQLite) across image builds, so rebuilds only
# recompile packages that actually changed.
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=1 go build -trimpath -ldflags="-s -w" -o main ./cmd/api

# Runtime stage
FROM alpine:3.21

# Deno is required by recent yt-dlp versions to solve YouTube's JavaScript
# challenges (nsig / PO tokens). The official Deno binary is built against
# glibc, so gcompat + libstdc++ let it run on Alpine's musl libc.
COPY --from=denoland/deno:bin-2.1.4 /deno /usr/local/bin/deno

RUN apk add --no-cache ffmpeg python3 curl gcompat libstdc++ \
    && curl -L https://github.com/yt-dlp/yt-dlp/releases/download/2026.03.17/yt-dlp -o /usr/local/bin/yt-dlp \
    && chmod +x /usr/local/bin/yt-dlp /usr/local/bin/deno

RUN addgroup -S appgroup && adduser -S appuser -G appgroup
WORKDIR /app

# The SQLite schema is embedded in the binary (services/backend/db/embed.go).
COPY --from=builder /build/main .

# Data directories are normally bind-mounted by docker compose; the loose
# permissions keep the image usable when run with an arbitrary --user.
RUN mkdir -p /app/data/downloads /app/data/db /app/data/processed /app/cache && \
    chown -R appuser:appgroup /app && \
    chmod -R 755 /app/data && \
    chmod -R 777 /app/data/db /app/data/processed /app/cache

ENV XDG_CACHE_HOME=/app/cache
USER appuser

EXPOSE 8080 8081

CMD ["./main"]
