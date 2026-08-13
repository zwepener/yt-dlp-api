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
	"os/exec"
	"strings"
	"sync"

	"urlasso-api/src/lib/env"
	"urlasso-api/src/lib/redis"
)

var (
	ctx = context.Background()
)

func main() {
	env.Init()

	redis.Init(
		env.REDIS_ADR,
		env.REDIS_USR,
		env.REDIS_PWD,
	)

	http.HandleFunc("/resolve", resolveHandler)
	http.HandleFunc("/ping", pingHandler)

	serverAddr := fmt.Sprintf("%s:%s", env.HOST, env.PORT)
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

	sem := make(chan struct{}, env.YTDLP_MCP)
	var waitGroup sync.WaitGroup
	mutex := sync.Mutex{}
	result := make(map[string]string)

	for _, url_ := range urls {
		url_, err := _cleanUrl(url_)
		if err != nil {
			continue
		}

		waitGroup.Add(1)
		go func(url_ string) {
			defer waitGroup.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			streamUrl, err := _resolveWithCache(url_)
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

func _resolveWithCache(url_ string) (string, error) {
	var err error

	cacheKey := "urlasso:" + _hashURL(url_)

	val, err := redis.RDB.Get(ctx, cacheKey).Result()
	if err == nil && strings.TrimSpace(val) != "" {
		return val, nil
	}

	streamUrl, err := _runYtDlp(url_)
	if err != nil {
		return "", err
	}

	err = redis.RDB.Set(ctx, cacheKey, streamUrl, env.CACHE_TTL).Err()
	if err != nil {
	}

	return streamUrl, nil
}

func _runYtDlp(url_ string) (string, error) {
	if url_ == "" {
		return "", errors.New("empty url")
	}

	cctx, cancel := context.WithTimeout(ctx, env.YTDLP_TMO)
	defer cancel()

	cmd := exec.CommandContext(
		cctx,
		env.YTDLP_CMD,
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

	scannerOut := bufio.NewScanner(stdout)
	var firstLine string
	for scannerOut.Scan() && scannerOut.Err() == nil {
		line := strings.TrimSpace(scannerOut.Text())
		if line != "" {
			firstLine = line
			break
		}
	}

	errBuf := new(strings.Builder)
	scannerErr := bufio.NewScanner(stderr)
	for scannerErr.Scan() && scannerErr.Err() == nil {
		errBuf.WriteString(scannerErr.Text())
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

/*
Removes unnecessary query parameters from a given url. Primarily for caching purposes.
Raises an error if the provided url string is empty or otherwise invalid.
*/
func _cleanUrl(rawUrl string) (string, error) {
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

func _hashURL(url_ string) string {
	hash := sha256.Sum256([]byte(url_))
	return hex.EncodeToString(hash[:])
}
