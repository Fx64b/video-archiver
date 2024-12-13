# Use an official Golang image as a base
FROM golang:1.23-alpine

# Install yt-dlp dependencies
ENV CGO_ENABLED=1
RUN apk add --no-cache gcc musl-dev curl ffmpeg python3

# Download yt-dlp and make it executable
RUN curl -L https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp -o /usr/local/bin/yt-dlp && \
    chmod +x /usr/local/bin/yt-dlp

WORKDIR /app

COPY go.mod ./
COPY go.sum ./


RUN go mod download

COPY . .

# Generate TypeScript files from Go structs
RUN go install github.com/gzuidhof/tygo@latest && \
    tygo generate && \
    mkdir -p /app/web/types && \
    mv generated/types/index.ts /app/web/types/index.ts

RUN go build -o main ./cmd/api

EXPOSE 8080
EXPOSE 8081

CMD ["./main"]
