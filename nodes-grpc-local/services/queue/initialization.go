package queue

import (
	"context"

	"github.com/redis/go-redis/v9"
)

func InitRedisConnection() *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:     REDIS_ADDRESS,
		Password: "",
		DB:       0,
	})

	_, err := rdb.Ping(context.Background()).Result()
	if err != nil {
		panic(err)
	}

	return rdb
}
