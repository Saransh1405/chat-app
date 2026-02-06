package redis

import (
	"chat-app/internal/config"

	"github.com/redis/go-redis/v9"
)

type Redis struct {
	Client *redis.Client
}

func NewRedis(cfg config.RedisConfig) (*Redis, error) {
	return &Redis{Client: redis.NewClient(&redis.Options{
		Addr:     cfg.Host + ":" + cfg.Port,
		Password: cfg.Password,
		DB:       cfg.DB,
	})}, nil
}
