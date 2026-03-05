package packages

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

func NewRedisClient(address string, password string, db int) (*redis.Client, error) {

	options := &redis.Options{
		Addr:     address,
		Password: password,
		DB:       db,
	}

	redisClient := redis.NewClient(options)

	contextWithTimeout, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := redisClient.Ping(contextWithTimeout).Err()
	if err != nil {
		return nil, fmt.Errorf("redis ping failed for %s: %w", address, err)
	}

	return redisClient, nil
}
