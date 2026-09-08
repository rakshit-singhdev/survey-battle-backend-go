package redis

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// releaseLockScript deletes the lock key only if it still holds the token
// that acquired it, so a lock never releases one it doesn't own (e.g. after
// its TTL expired and another caller acquired it in the meantime).
var releaseLockScript = redis.NewScript(`
if redis.call("get", KEYS[1]) == ARGV[1] then
	return redis.call("del", KEYS[1])
end
return 0
`)

func newLockToken() (string, error) {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}

	return hex.EncodeToString(buf), nil
}

func (r *RedisStore) AcquireGameLock(
	ctx context.Context,
	sessionID string,
	ttl time.Duration,
) (string, error) {

	token, err := newLockToken()
	if err != nil {
		return "", err
	}

	ok, err := r.Client.SetNX(
		ctx,
		GameLockKey(sessionID),
		token,
		ttl,
	).Result()

	if err != nil {
		return "", err
	}

	if !ok {
		return "", nil
	}

	return token, nil
}

func (r *RedisStore) ReleaseGameLock(
	ctx context.Context,
	sessionID string,
	token string,
) error {

	if token == "" {
		return nil
	}

	return releaseLockScript.Run(
		ctx,
		r.Client,
		[]string{GameLockKey(sessionID)},
		token,
	).Err()
}

func (r *RedisStore) WithGameLock(
	ctx context.Context,
	sessionID string,
	fn func() error,
) error {

	token, err := r.AcquireGameLock(
		ctx,
		sessionID,
		r.GameLockTTL,
	)

	if err != nil {
		return err
	}

	if token == "" {
		return fmt.Errorf("game state is busy, please retry")
	}

	defer func() {
		_ = r.ReleaseGameLock(
			ctx,
			sessionID,
			token,
		)
	}()

	return fn()
}

func (r *RedisStore) TryBuzzerLock(
	ctx context.Context,
	sessionID string,
	roundID string,
	teamID string,
	ttl time.Duration,
) (bool, error) {

	return r.Client.SetNX(
		ctx,
		BuzzerLockKey(sessionID, roundID),
		teamID,
		ttl,
	).Result()
}

func (r *RedisStore) ReleaseBuzzerLock(
	ctx context.Context,
	sessionID string,
	roundID string,
) error {

	return r.Client.Del(
		ctx,
		BuzzerLockKey(sessionID, roundID),
	).Err()
}
