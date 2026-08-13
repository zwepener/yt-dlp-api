package env

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
)

var (
	HOST string
	PORT string

	YTDLP_CMD string
	YTDLP_TMO time.Duration
	YTDLP_MCP int

	REDIS_ADR string
	REDIS_USR string
	REDIS_PWD string
	CACHE_TTL time.Duration
)

func Init() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, relying on system environment variables!")
	}

	HOST = getStrEnv("SERVER_HOST", "0.0.0.0")
	PORT = getStrEnv("SERVER_PORT", "8080")

	YTDLP_CMD = getStrEnv("YTDLP_CMD", "yt-dlp")
	YTDLP_TMO = getDurEnv("YTDLP_TMO", 15*time.Second)
	YTDLP_MCP = getIntEnv("YTDLP_MCP", 2)

	REDIS_ADR = getStrEnv("REDIS_ADR", "localhost:6379")
	REDIS_USR = getStrEnv("REDIS_USR", "default")
	REDIS_PWD = getStrEnv("REDIS_PWD", "")
	CACHE_TTL = getDurEnv("CACHE_TTL", 6*time.Hour)
}

func getStrEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getDurEnv(key string, fallback time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return fallback
}

func getIntEnv(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		var i int
		if _, err := fmt.Sscanf(v, "%d", &i); err == nil {
			return i
		}
	}
	return fallback
}
