package redis

import (
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
}
