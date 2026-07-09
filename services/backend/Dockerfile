# Build stage: compile the Go binary.
# The SQLite driver is pure Go (modernc.org/sqlite), so no C toolchain is
# needed and the binary is fully static — this is what keeps builds fast.
FROM golang:1.25-alpine AS builder

WORKDIR /build

# Resolve modules first so they are cached independently of source changes.
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

COPY . .

# The cache mounts persist the module cache and compiled objects across image
# builds, so rebuilds only recompile packages that actually changed.
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o main ./cmd/api

# Dev stage: hot reload via Air with the source bind-mounted by
# docker-compose.dev.yml. Media tools match the runtime image so downloads
# and transcodes work identically in development.
FROM golang:1.25-alpine AS dev

COPY --from=denoland/deno:bin-2.1.4 /deno /usr/local/bin/deno
RUN apk add --no-cache ffmpeg python3 curl gcompat libstdc++ \
    && curl -L https://github.com/yt-dlp/yt-dlp/releases/download/2026.03.17/yt-dlp -o /usr/local/bin/yt-dlp \
    && chmod +x /usr/local/bin/yt-dlp /usr/local/bin/deno

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 go install github.com/air-verse/air@v1.61.7

WORKDIR /app
ENV CGO_ENABLED=0
CMD ["air"]

# Runtime stage
FROM alpine:3.24 AS runner

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

EXPOSE 8080

CMD ["./main"]
