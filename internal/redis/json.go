package redis

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
)

func (r *RedisStore) ReadJSON(
	ctx context.Context,
	key string,
	dest any,
) error {
	value, err := r.Client.Get(ctx, key).Result()

	if err == redis.Nil {
		return nil
	}

	if err != nil {
		return err
	}

	return json.Unmarshal([]byte(value), dest)
}

func (r *RedisStore) WriteJSON(
	ctx context.Context,
	key string,
	value any,
	ttl time.Duration,
) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	return r.Client.Set(
		ctx,
		key,
		data,
		ttl,
	).Err()
}

func (r *RedisStore) Delete(
	ctx context.Context,
	key string,
) error {
	return r.Client.Del(ctx, key).Err()
}
