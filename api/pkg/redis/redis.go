package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/chattycathy/api/pkg/logger"
	"github.com/redis/go-redis/v9"
)

var Client *redis.Client

type Config struct {
	Host     string
	Port     string
	Password string
	DB       int
}

func Connect(cfg *Config) error {
	addr := fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)

	Client = redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := Client.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("failed to connect to Redis: %w", err)
	}

	logger.Info().Str("addr", addr).Msg("Redis connected successfully")
	return nil
}

func Close() error {
	if Client != nil {
		return Client.Close()
	}
	return nil
}

func IsConnected() bool {
	if Client == nil {
		return false
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	return Client.Ping(ctx).Err() == nil
}
