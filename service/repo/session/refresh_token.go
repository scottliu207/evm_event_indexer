package session

import (
	"context"
	"evm_event_indexer/internal/config"
	"evm_event_indexer/internal/session"
	"evm_event_indexer/internal/storage"
	"fmt"
)

func newUserRTKey(userID int64) string {
	return fmt.Sprintf("user_rt:%d", userID)
}

func CreateRefreshToken(ctx context.Context, userID int64) (string, error) {

	rt, err := session.NewOpaqueToken()
	if err != nil {
		return "", fmt.Errorf("failed to generate opaque token: %w", err)
	}

	redis, err := storage.GetRedis(config.RedisUser)
	if err != nil {
		return "", fmt.Errorf("failed to get redis: %w", err)
	}

	if err := redis.Set(ctx, newUserRTKey(userID), rt, config.Get().Session.RTExpiration).Err(); err != nil {
		return "", fmt.Errorf("failed to set refresh token: %w", err)
	}

	return rt, nil
}
