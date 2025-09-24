FROM golang:1.25-alpine AS builder
WORKDIR /app

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build ./src/main.go

FROM python:3.13-alpine AS runner
WORKDIR /app

RUN pip install --no-cache-dir yt-dlp

COPY --from=builder /app/main /app/api
RUN chmod +x /app/api

EXPOSE 8080

ENTRYPOINT ["/app/api"]
