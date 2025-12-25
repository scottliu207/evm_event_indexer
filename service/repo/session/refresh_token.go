package session

import (
	"context"
	"evm_event_indexer/internal/config"
	"evm_event_indexer/internal/session"
	"fmt"

	"github.com/redis/go-redis/v9"
)

func newUserRTKey(rt string) string {
	return fmt.Sprintf("user_rt:%s", rt)
}

func CreateRefreshToken(ctx context.Context, client *redis.Client, userID int64) (string, error) {

	rt, err := session.NewOpaqueToken()
	if err != nil {
		return "", fmt.Errorf("failed to generate opaque token: %w", err)
	}

	if err := client.Set(ctx, newUserRTKey(rt), userID, config.Get().Session.RTExpiration).Err(); err != nil {
		return "", fmt.Errorf("failed to set refresh token: %w", err)
	}

	return rt, nil
}

func DeleteRefreshToken(ctx context.Context, client *redis.Client, rt string) error {
	if err := client.Del(ctx, newUserRTKey(rt)).Err(); err != nil {
		return fmt.Errorf("failed to delete refresh token: %w", err)
	}

	return nil
}

// GetUserIDByRefreshToken gets a userID by refresh token, return 0 if not found.
func GetUserIDByRefreshToken(ctx context.Context, client *redis.Client, rt string) (int64, error) {
	userID, err := client.Get(ctx, newUserRTKey(rt)).Int64()
	if err != nil && err != redis.Nil {
		return 0, fmt.Errorf("failed to get refresh token: %w", err)
	}

	return userID, nil
}
