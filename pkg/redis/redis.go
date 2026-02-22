package redis

import (
	"context"
	"log"
	"time"

	"github.com/redis/go-redis/v9"

	"game-server/internal/config"
)

var Client *redis.Client

func InitRedis() error {
	cfg := config.GlobalConfig.Redis

	Client = redis.NewClient(&redis.Options{
		Addr:         cfg.Addr(),
		Password:     cfg.Password,
		DB:           cfg.DB,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		PoolSize:     100,
		MinIdleConns: 10,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := Client.Ping(ctx).Err(); err != nil {
		return err
	}

	log.Println("Redis connected successfully")
	return nil
}

func CloseRedis() {
	if Client != nil {
		Client.Close()
	}
}
