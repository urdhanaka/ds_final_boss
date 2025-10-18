package config

import (
	"context"

	"github.com/redis/go-redis/v9"
)

const (
	BE_REDIS_ADDRESS = "localhost:6379"
)

func NewRedisConnection() *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:     BE_REDIS_ADDRESS,
		Password: "",
		DB:       0,
	})

	_, err := rdb.Ping(context.Background()).Result()
	if err != nil {
		panic(err)
	}

	return rdb
}
