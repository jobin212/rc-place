package main

import (
	"context"
	"os"

	"github.com/go-redis/redis/v8"
)

var redisClient *redis.Client

// newRedisClient creates a connection to redis
func setupRedisClient() error {
	ctx := context.Background()
	redisClient = redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_HOST"),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       0, // default DB
	})

	return redisClient.Ping(ctx).Err()
}

func getColorsFromByte(b byte) (firstColor, secondColor int) {
	return (int(b >> 4)), (int(b & 15))
}
