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

	REDIS_ON  bool
	REDIS_ADR string
	REDIS_USR string
	REDIS_PWD string
	CACHE_TTL time.Duration
)

func Init() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, relying on system environment variables!")
	}

	HOST, _ = getStrEnv("SERVER_HOST", "0.0.0.0")
	PORT, _ = getStrEnv("SERVER_PORT", "8080")

	YTDLP_CMD, _ = getStrEnv("YTDLP_CMD", "yt-dlp")
	YTDLP_TMO, _ = getDurEnv("YTDLP_TMO", 15*time.Second)
	YTDLP_MCP, _ = getIntEnv("YTDLP_MCP", 2)

	REDIS_ADR, REDIS_ON = getStrEnv("REDIS_ADR", "localhost:6379")
	REDIS_USR, _ = getStrEnv("REDIS_USR", "default")
	REDIS_PWD, _ = getStrEnv("REDIS_PWD", "")
	CACHE_TTL, _ = getDurEnv("CACHE_TTL", 6*time.Hour)
}

func getStrEnv(key string, fallback string) (string, bool) {
	value, exists := os.LookupEnv(key)
	if value != "" {
		return value, exists
	}
	return fallback, exists
}

func getDurEnv(key string, fallback time.Duration) (time.Duration, bool) {
	value, exists := os.LookupEnv(key)
	if value != "" {
		if d, err := time.ParseDuration(value); err == nil {
			return d, exists
		}
	}
	return fallback, exists
}

func getIntEnv(key string, fallback int) (int, bool) {
	value, exists := os.LookupEnv(key)
	if value != "" {
		var i int
		if _, err := fmt.Sscanf(value, "%d", &i); err == nil {
			return i, exists
		}
	}
	return fallback, exists
}
