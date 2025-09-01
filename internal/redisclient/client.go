package redisclient

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/redis/go-redis/v9"
)

var (
	redisClient *redis.Client
	once        sync.Once
)

func GetRedisClient() *redis.Client {
	once.Do(func() {
		rdb, err := NewRedisClient(context.Background(), &redis.Options{
			Addr:     "localhost:6379",
			Password: "", // no password set
			DB:       0,  // use default DB
		})
		if err != nil {
			log.Fatalf("Failed to create Redis client: %v", err)
		}
		redisClient = rdb
	})
	return redisClient
}

func NewRedisClient(ctx context.Context, options *redis.Options) (*redis.Client, error) {
	rdb := redis.NewClient(options)

	// Ping the server to ensure the connection is successful.
	// This check is a good practice to fail fast if the server is not available.
	pong, err := rdb.Ping(ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}
	fmt.Println("Connected to Redis successfully:", pong)

	return rdb, nil
}
