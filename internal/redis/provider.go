package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"survey-battle-backend-go/internal/config"
)

type RedisStore struct {
	Client *redis.Client

	GameStateTTL time.Duration
	GameLockTTL  time.Duration
}

func Connect(ctx context.Context, url string, cfg config.RedisConfig) (*RedisStore, error) {
	opts, err := redis.ParseURL(url)
	if err != nil {
		return nil, fmt.Errorf("parse Redis URL: %w", err)
	}

	client := redis.NewClient(opts)

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := client.Ping(pingCtx).Err(); err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("ping Redis: %w", err)
	}

	return &RedisStore{
		Client:       client,
		GameStateTTL: cfg.GameStateTTL,
		GameLockTTL:  cfg.GameLockTTL,
	}, nil
}

func (r *RedisStore) Disconnect() error {
	return r.Client.Close()
}
