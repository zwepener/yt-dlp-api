FROM golang:1.25-alpine AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o ./api ./src


FROM debian:bookworm-slim AS ytdlp
RUN apt-get update && apt-get install -y --no-install-recommends curl ca-certificates && \
    curl -L "https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp_linux" -o /yt-dlp && \
    chmod +x /yt-dlp


FROM gcr.io/distroless/base-debian12 AS runner
WORKDIR /app

COPY --from=builder /app/api ./api
COPY --from=ytdlp /yt-dlp /usr/local/bin/yt-dlp

# Additional libraries required by yt-dlp
COPY --from=ytdlp /usr/lib/x86_64-linux-gnu/libz.so.1 /usr/lib/x86_64-linux-gnu/libz.so.1

EXPOSE 8080

ENTRYPOINT ["./api"]
