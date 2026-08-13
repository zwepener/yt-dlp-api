package redis

import (
	"context"
	"log"
	"time"

	"github.com/redis/go-redis/v9"

	"urlasso-api/src/lib/env"
)

var RDB *redis.Client

func Init(address string, username string, password string) {
	if !env.REDIS_ON {
		log.Println("REDIS_ADR environment variable is undefined. Caching will be disabled.")
		return
	}

	RDB = redis.NewClient(&redis.Options{
		Addr:     address,
		Username: username,
		Password: password,
		DB:       0,

		DialTimeout:  time.Second,
		ReadTimeout:  time.Second,
		WriteTimeout: time.Second,

		MaxRetries:      10,
		MinRetryBackoff: 1 * time.Second,
		MaxRetryBackoff: 10 * time.Second,
	})

	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			retryPing(RDB, 5)
		}
	}()
}

func retryPing(rdb *redis.Client, maxRetries int) {
	timeout := 2 * time.Second
	for i := range maxRetries {
		ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
		err := rdb.Ping(ctx).Err()
		cancel()

		if err == nil {
			return
		}
		log.Printf("Redis ping attempt %d failed: %v", i+1, err)

		time.Sleep(timeout)
	}
	log.Println("Redis ping failed after retries")
}
