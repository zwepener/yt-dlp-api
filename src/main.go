package main

import (
	"bufio"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
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
	redisAddr      string
	cacheTTL       time.Duration
	serverAddr     string
	ytDlpCmd       string
	perCallTimeout time.Duration
	maxConcurrency int
)

func init_env() {
	log.Println("loading environment variables . . .")
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, relying on system environment variables!")
	}

	redisAddr = getEnv("REDIS_ADDR", "localhost:6379")
	cacheTTL = getEnvDuration("CACHE_TTL", 6*time.Hour)
	serverAddr = getEnv("SERVER_ADDR", ":8080")
	ytDlpCmd = getEnv("YTDLP_CMD", "yt-dlp")
	perCallTimeout = getEnvDuration("YTDLP_TIMEOUT", 15*time.Second)
	maxConcurrency = getEnvInt("MAX_CONCURRENCY", 8)
}

func main() {
	init_env()

	log.Println("connecting to redis client . . .")
	redisClient = redis.NewClient(&redis.Options{
		Addr: redisAddr,
		DB:   0,
	})

	if err := redisClient.Ping(ctx).Err(); err != nil {
		redisClient = nil
		log.Printf("failed to connect to redis at %s: %v", redisAddr, err)
		log.Print("Caching will be disabled.")
	}

	http.HandleFunc("/resolve", resolveHandler)
	http.HandleFunc("/heartbeat", pingHandler)

	log.Printf("server listening on %s", serverAddr)
	if err := http.ListenAndServe(serverAddr, nil); err != nil {
		log.Fatalf("server exited: %v", err)
	}
}

func pingHandler(res http.ResponseWriter, req *http.Request) {
	res.WriteHeader(http.StatusOK)
}

func resolveHandler(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(res, "only POST allowed", http.StatusMethodNotAllowed)
		return
	}

	var urls []string

	decoder := json.NewDecoder(req.Body)
	if err := decoder.Decode(&urls); err != nil {
		http.Error(res, "invalid JSON body; expected array of strings", http.StatusBadRequest)
		return
	}

	if len(urls) == 0 {
		res.Header().Set("Content-Type", "application/json")
		res.Write([]byte("{}"))
		return
	}

	sem := make(chan struct{}, maxConcurrency) // sem = semaphore
	var waitGroup sync.WaitGroup
	mutex := sync.Mutex{}
	result := make(map[string]string)

	for _, url_ := range urls {
		url_, err := cleanUrl(url_)
		if err != nil {
			continue
		}

		waitGroup.Add(1)
		go func(url_ string) {
			defer waitGroup.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			streamUrl, err := resolveWithCache(url_)
			if err != nil {
				log.Printf("could not resolve %s: %v", url_, err)
				return
			}

			mutex.Lock()
			result[url_] = streamUrl
			mutex.Unlock()
		}(url_)
	}

	waitGroup.Wait()

	res.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(res)
	if err := enc.Encode(result); err != nil {
		log.Printf("failed to encode result: %v", err)
	}
}

func resolveWithCache(url_ string) (string, error) {
	cacheKey := "yt-dlp:" + hashURL(url_)

	log.Print(cacheKey)

	if redisClient != nil {
		val, err := redisClient.Get(ctx, cacheKey).Result()
		if err == nil && strings.TrimSpace(val) != "" {
			return val, nil
		}
	}

	streamURL, err := runYtDlp(url_)
	if err != nil {
		return "", err
	}

	if redisClient != nil {
		err = redisClient.Set(ctx, cacheKey, streamURL, cacheTTL).Err()
		if err != nil {
			log.Printf("warning: failed to set cache for %s: %v", url_, err)
		}
	}

	return streamURL, nil
}

func runYtDlp(url_ string) (string, error) {
	if url_ == "" {
		return "", errors.New("empty url")
	}

	cctx, cancel := context.WithTimeout(ctx, perCallTimeout)
	defer cancel()

	cmd := exec.CommandContext(
		cctx,
		ytDlpCmd,
		"--get-url", "--no-playlist", "--no-warnings", "--no-cache-dir",
		url_,
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

/**
 * Removes unnecessary query parameters from a given url. Primarily for caching purposes
 * Raises an error if the provided url string is empty or otherwise invalid.
 */
func cleanUrl(rawUrl string) (string, error) {
	rawUrl = strings.TrimSpace(rawUrl)
	if rawUrl == "" {
		return "", errors.New("url is empty")
	}

	url_, err := url.Parse(rawUrl)
	if err != nil {
		return "", err
	}

	junk_params := map[string]bool{
		"igsh": true, "si": true, "mibextid": true,
	}

	q := url_.Query()

	for param := range q {
		if junk_params[param] {
			q.Del(param)
		}
	}

	url_.RawQuery = q.Encode()

	return url_.String(), nil
}

func hashURL(url_ string) string {
	hash := sha256.Sum256([]byte(url_))
	return hex.EncodeToString(hash[:])
}
