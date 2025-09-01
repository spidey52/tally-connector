package main

import (
	"context"
	"log"
	"tally-connector/internal/redisclient"

	"github.com/redis/go-redis/v9"
)

func ProcessImportQueue() {

	client, err := redisclient.NewRedisClient(context.Background(), &redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	if err != nil {
		log.Fatalf("Failed to create Redis client: %v", err)
	}

	defer client.Close()

	for {
		val := client.BRPop(context.Background(), 0, redisclient.ImportQueueKey).Val()

		if len(val) < 2 {
			log.Println("Invalid value found in import queue")
			continue
		}

		log.Println("Processing import:", val[0], val[1])
		ImportAll(val[1])

	}

}
