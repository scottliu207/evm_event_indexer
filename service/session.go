package service

import (
	"context"
	"evm_event_indexer/internal/config"
	"evm_event_indexer/internal/errors"
	internalSession "evm_event_indexer/internal/session"
	"evm_event_indexer/internal/storage"
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

	client, err := storage.GetRedis(config.RedisUser)
	if err != nil {
		return "", errors.ErrInternalServerError.Wrap(err, "failed to get redis")
	}

	rt, err := session.CreateRefreshToken(ctx, client, userID)
	if err != nil {
		return "", errors.ErrInternalServerError.Wrap(err, "failed to create refresh token")
	}

	return rt, nil
}

// VerifyUserAT verifies a user access token
func VerifyUserAT(at string) (int64, error) {
	userID, err := internalSession.NewJWT(config.Get().Session.JWTSecret).VerifyToken(at)
	if err != nil {
		return 0, errors.ErrInvalidCredentials.Wrap(err)
	}

	if userID == 0 {
		return 0, errors.ErrInvalidCredentials.New()
	}

	return userID, nil
}

// GetUserByRT gets a userID by refresh token, return InvalidCredentialsError if the refresh token is invalid.
func GetUserIDByRT(ctx context.Context, rt string) (int64, error) {
	client, err := storage.GetRedis(config.RedisUser)
	if err != nil {
		return 0, errors.ErrInternalServerError.Wrap(err, "failed to get redis")
	}

	userID, err := session.GetUserIDByRefreshToken(ctx, client, rt)
	if err != nil {
		return 0, errors.ErrInternalServerError.Wrap(err, "failed to get user id by refresh token")
	}

	if userID == 0 {
		return 0, errors.ErrInvalidCredentials.New("invalid refresh token")
	}

	return userID, nil
}

// DeleteUserRT deletes a user refresh token
func DeleteUserRT(ctx context.Context, rt string) error {

	if rt == "" {
		return errors.ErrApiInvalidParam.New("refresh token is required")
	}

	client, err := storage.GetRedis(config.RedisUser)
	if err != nil {
		return errors.ErrInternalServerError.Wrap(err, "failed to get redis")
	}

	err = session.DeleteRefreshToken(ctx, client, rt)
	if err != nil {
		return errors.ErrInternalServerError.Wrap(err, "failed to delete refresh token")
	}

	return nil

}
