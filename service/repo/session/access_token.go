package session

import (
	"context"
	"evm_event_indexer/internal/config"
	"evm_event_indexer/internal/session"
	"fmt"
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
