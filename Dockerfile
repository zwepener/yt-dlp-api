FROM golang:1.25-alpine AS builder
WORKDIR /app

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .
# RUN CGO_ENABLED=0 GOOS=linux go build ./src/main.go -trimpath -ldflags="-s -w" -o /app/api
RUN CGO_ENABLED=0 GOOS=linux go build ./src/main.go


FROM alpine:latest as base


FROM base as yt-dlp
WORKDIR /app

RUN apk add --no-cache curl

RUN curl -L https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp_linux -o yt-dlp


FROM base AS runner
WORKDIR /app

COPY --from=yt-dlp /app/yt-dlp /usr/bin/yt-dlp
RUN chmod +x /usr/bin/yt-dlp

COPY --from=builder /app/main /app/api
RUN chmod +x /app/api

EXPOSE 8080

ENTRYPOINT ["/app/api"]
