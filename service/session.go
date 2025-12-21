package service

import (
	"context"
	"evm_event_indexer/internal/config"
	"evm_event_indexer/internal/errors"
	internalSession "evm_event_indexer/internal/session"
	"evm_event_indexer/service/repo/session"
)

func CreateUserAT(ctx context.Context, userID int64) (string, error) {

	if userID == 0 {
		return "", errors.ErrApiInvalidParam.New("user id is 0")
	}

	at, err := internalSession.NewJWT(config.Get().Session.JWTSecret).GenerateToken(userID, config.Get().Session.ATExpiration)
	if err != nil {
		return "", errors.ErrInternalServerError.Wrap(err, "failed to generate token")
	}

	return at, nil
}

// CreateUserRT creates a user refresh token
func CreateUserRT(ctx context.Context, userID int64) (string, error) {
	if userID == 0 {
		return "", errors.ErrApiInvalidParam.New("user id is 0")
	}

	rt, err := session.CreateRefreshToken(ctx, userID)
	if err != nil {
		return "", errors.ErrInternalServerError.Wrap(err, "failed to create refresh token")
	}

	return rt, nil
}

// VerifyUserAT verifies a user access token
func VerifyUserAT(at string) (int64, error) {
	userID, err := internalSession.NewJWT(config.Get().Session.JWTSecret).VerifyToken(at)
	if err != nil {
		return 0, errors.ErrInternalServerError.Wrap(err, "failed to verify access token")
	}

	if userID == 0 {
		return 0, errors.ErrInvalidCredentials.New()
	}

	return userID, nil
}
