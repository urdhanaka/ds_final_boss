package config

import (
	"context"
	consts "nodes-grpc-be/const"

	"github.com/redis/go-redis/v9"
)

func NewRedisConnection() *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:     consts.REDIS_ADDRESS,
		Password: "",
		DB:       0,
	})

	_, err := rdb.Ping(context.Background()).Result()
	if err != nil {
		panic(err)
	}

	return rdb
}
