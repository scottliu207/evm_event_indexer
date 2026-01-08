package session

import (
	"context"
	"evm_event_indexer/internal/config"
	"evm_event_indexer/internal/session"
	"fmt"
	"strconv"

	"github.com/redis/go-redis/v9"
)

// CreateAccessToken creates a user access token
func CreateAccessToken(ctx context.Context, userID int64) (string, error) {
	at, err := session.NewJWT(config.Get().Session.JWTSecret).GenerateToken(userID, config.Get().Session.ATExpiration)
	if err != nil {
		return "", fmt.Errorf("failed to generate token: %w", err)
	}

	return at, nil
}

// VerifyAccessToken verifies a user access token
func VerifyAccessToken(at string) (int64, error) {
	return session.NewJWT(config.Get().Session.JWTSecret).VerifyToken(at)
}

func newUserRTKey(rt string) string {
	return fmt.Sprintf("user_rt:rt:%s", rt)
}

func newUserRTUserIDKey(userID int64) string {
	return fmt.Sprintf("user_rt:user_id:%d", userID)
}

// CreateRefreshToken creates a user refresh token
func CreateRefreshToken(ctx context.Context, client *redis.Client, userID int64) (string, error) {

	rt, err := session.NewOpaqueToken()
	if err != nil {
		return "", fmt.Errorf("failed to generate opaque token: %w", err)
	}

	_, err = client.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
		pipe.Set(ctx, newUserRTKey(rt), userID, config.Get().Session.RTExpiration)
		pipe.Set(ctx, newUserRTUserIDKey(userID), rt, config.Get().Session.RTExpiration)
		return nil
	})
	if err != nil {
		return "", fmt.Errorf("failed to create refresh token: %w", err)
	}

	return rt, nil
}

// DeleteRefreshTokenByUserID delete a user refresh token by user id after retreive refresh token
func DeleteRefreshTokenByUserID(ctx context.Context, client *redis.Client, userID int64) error {

	rt, err := client.Get(ctx, newUserRTUserIDKey(userID)).Result()
	if err != nil && err != redis.Nil {
		return fmt.Errorf("failed to delete user id: %w", err)
	}

	// if refresh token is not found, skip delete
	if rt == "" {
		return nil
	}

	_, err = client.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
		pipe.Del(ctx, newUserRTKey(rt))
		pipe.Del(ctx, newUserRTUserIDKey(userID))
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to delete refresh token: %w", err)
	}
	return nil
}

// DeleteRefreshTokenByRT delete a user refresh token by refresh token after retreive user id
func DeleteRefreshTokenByRT(ctx context.Context, client *redis.Client, rt string) error {
	userIDStr, err := client.Get(ctx, newUserRTKey(rt)).Result()
	if err != nil && err != redis.Nil {
		return fmt.Errorf("failed to delete user id: %w", err)
	}

	if userIDStr == "" {
		return nil
	}

	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		return fmt.Errorf("failed to parse user id: %w", err)
	}

	_, err = client.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
		pipe.Del(ctx, newUserRTKey(rt))
		pipe.Del(ctx, newUserRTUserIDKey(userID))
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to delete refresh token: %w", err)
	}

	return nil
}

// GetUserIDByRT gets a user id by refresh token
func GetUserIDByRT(ctx context.Context, client *redis.Client, rt string) (int64, error) {
	userID, err := client.Get(ctx, newUserRTKey(rt)).Int64()
	if err != nil && err != redis.Nil {
		return 0, fmt.Errorf("failed to get user id: %w", err)
	}

	return userID, nil
}
