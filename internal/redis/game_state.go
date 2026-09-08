package redis

import (
	"context"
	"encoding/json"
)

func (r *RedisStore) WriteGameLiveState(
	ctx context.Context,
	sessionID string,
	state any,
) error {
	return r.WriteJSON(
		ctx,
		GameStateKey(sessionID),
		state,
		r.GameStateTTL,
	)
}

func (r *RedisStore) ReadGameLiveState(
	ctx context.Context,
	sessionID string,
	dest any,
) error {
	return r.ReadJSON(
		ctx,
		GameStateKey(sessionID),
		dest,
	)
}

func (r *RedisStore) WriteGameLiveStateIfAbsent(
	ctx context.Context,
	sessionID string,
	state any,
) (bool, error) {

	data, err := json.Marshal(state)
	if err != nil {
		return false, err
	}

	ok, err := r.Client.SetNX(
		ctx,
		GameStateKey(sessionID),
		data,
		r.GameStateTTL,
	).Result()

	return ok, err
}
