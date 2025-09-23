package main

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
)

var (
	redisClient *redis.Client
	ctx         = context.Background()
)

var (
	redisAddr      = getEnv("REDIS_ADDR", "redis:6379")
	redisPassword  = getEnv("REDIS_PASSWORD", "")
	cacheTTL       = getEnvDuration("CACHE_TTL", 24*time.Hour)
	serverAddr     = getEnv("SERVER_ADDR", ":8080")
	ytDlpCmd       = getEnv("YTDLP_CMD", "yt-dlp")
	perCallTimeout = getEnvDuration("YTDLP_TIMEOUT", 15*time.Second)
	maxConcurrency = getEnvInt("MAX_CONCURRENCY", 8)
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, relying on system environment variables!")
	}

	redisClient = redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: redisPassword,
		DB:       0,
	})

	if err := redisClient.Ping(ctx).Err(); err != nil {
		log.Fatalf("failed to connect to redis at %s: %v", redisAddr, err)
	}

	http.HandleFunc("/resolve", resolveHandler)

	log.Printf("server listening on %s", serverAddr)
	if err := http.ListenAndServe(serverAddr, nil); err != nil {
		log.Fatalf("server exited: %v", err)
	}
}

func resolveHandler(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(res, "only POST allowed", http.StatusMethodNotAllowed)
		return
	}

	var urls []string
	dec := json.NewDecoder(req.Body)
	if err := dec.Decode(&urls); err != nil {
		http.Error(res, "invalid JSON body; expected array of strings", http.StatusBadRequest)
		return
	}
	if len(urls) == 0 {
		res.Header().Set("Content-Type", "application/json")
		res.Write([]byte("{}"))
		return
	}

	sem := make(chan struct{}, maxConcurrency)
	var waitGroup sync.WaitGroup
	mu := sync.Mutex{}
	result := make(map[string]string)

	for _, u := range urls {
		u := strings.TrimSpace(u)
		if u == "" {
			continue
		}

		waitGroup.Add(1)
		go func(url string) {
			defer waitGroup.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			streamURL, err := resolveWithCache(url)
			if err != nil {
				log.Printf("could not resolve %s: %v", url, err)
				return
			}
			mu.Lock()
			result[url] = streamURL
			mu.Unlock()
		}(u)
	}

	waitGroup.Wait()

	res.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(res)
	if err := enc.Encode(result); err != nil {
		log.Printf("failed to encode result: %v", err)
	}
}

func resolveWithCache(url string) (string, error) {
	cacheKey := "yt-dlp:" + url

	val, err := redisClient.Get(ctx, cacheKey).Result()
	if err == nil && strings.TrimSpace(val) != "" {
		return val, nil
	}

	streamURL, err := runYtDlp(url)
	if err != nil {
		return "", err
	}

	err = redisClient.Set(ctx, cacheKey, streamURL, cacheTTL).Err()
	if err != nil {
		log.Printf("warning: failed to set cache for %s: %v", url, err)
	}

	return streamURL, nil
}

func runYtDlp(url string) (string, error) {
	if url == "" {
		return "", errors.New("empty url")
	}

	cctx, cancel := context.WithTimeout(ctx, perCallTimeout)
	defer cancel()

	cmd := exec.CommandContext(
		cctx,
		ytDlpCmd,
		"--get-url", "--no-playlist", "--no-warnings", "--no-cache-dir",
		url,
	)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return "", fmt.Errorf("failed to get stdout pipe: %w", err)
	}
	stderr, _ := cmd.StderrPipe()

	if err := cmd.Start(); err != nil {
		return "", fmt.Errorf("failed to start yt-dlp: %w", err)
	}

	scanner := bufio.NewScanner(stdout)
	var firstLine string
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			firstLine = line
			break
		}
	}

	errBuf := new(strings.Builder)
	serr := bufio.NewScanner(stderr)
	for serr.Scan() {
		errBuf.WriteString(serr.Text())
		errBuf.WriteByte('\n')
	}

	if err := cmd.Wait(); err != nil {
		stderrText := strings.TrimSpace(errBuf.String())
		if stderrText != "" {
			return "", fmt.Errorf("yt-dlp failed: %v; stderr: %s", err, stderrText)
		}
		return "", fmt.Errorf("yt-dlp failed: %w", err)
	}

	if firstLine == "" {
		return "", errors.New("no streaming url returned by yt-dlp")
	}

	return firstLine, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvDuration(key string, fallback time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		var i int
		_, err := fmt.Sscanf(v, "%d", &i)
		if err == nil {
			return i
		}
	}
	return fallback
}
