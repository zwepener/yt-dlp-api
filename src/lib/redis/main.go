package redis

import (
	"time"

	"github.com/redis/go-redis/v9"
)

var RDB *redis.Client

func Init(address string, username string, password string) {
	RDB = redis.NewClient(&redis.Options{
		Addr:     address,
		Username: username,
		Password: password,
		DB:       0,

		DialTimeout:  time.Second,
		ReadTimeout:  time.Second,
		WriteTimeout: time.Second,

		MaxRetries:      10,
		MinRetryBackoff: 10 * time.Millisecond,
		MaxRetryBackoff: 5 * time.Second,
	})
}
