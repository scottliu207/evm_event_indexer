package service

import (
	"context"
	"evm_event_indexer/internal/config"
	"evm_event_indexer/internal/errors"
	"evm_event_indexer/internal/storage"
	"evm_event_indexer/service/repo/session"
)

// CreateUserAT creates a user access token
func CreateUserAT(ctx context.Context, userID int64) (string, error) {

	if userID == 0 {
		return "", errors.ErrApiInvalidParam.New("user id is 0")
	}

	at, err := session.CreateAccessToken(ctx, userID)
	if err != nil {
		return "", errors.ErrInternalServerError.Wrap(err, "failed to create access token")
	}

	return at, nil
}

// VerifyUserAT verifies a user access token
func VerifyUserAT(at string) (int64, error) {
	userID, err := session.VerifyAccessToken(at)
	if err != nil {
		return 0, errors.ErrInvalidCredentials.Wrap(err)
	}

	if userID == 0 {
		return 0, errors.ErrInvalidCredentials.New()
	}

	return userID, nil
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

// GetUserByRT gets a userID by refresh token, return InvalidCredentialsError if the refresh token is invalid.
func GetUserIDByRT(ctx context.Context, rt string) (int64, error) {
	client, err := storage.GetRedis(config.RedisUser)
	if err != nil {
		return 0, errors.ErrInternalServerError.Wrap(err, "failed to get redis")
	}

	userID, err := session.GetUserIDByRT(ctx, client, rt)
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

	if err = session.DeleteRefreshTokenByRT(ctx, client, rt); err != nil {
		return errors.ErrInternalServerError.Wrap(err, "failed to delete refresh token")
	}

	return nil

}

// DeleteUserRT deletes a user refresh token
func DeleteUserRTByUserID(ctx context.Context, userID int64) error {

	if userID == 0 {
		return errors.ErrApiInvalidParam.New("user id is 0")
	}

	client, err := storage.GetRedis(config.RedisUser)
	if err != nil {
		return errors.ErrInternalServerError.Wrap(err, "failed to get redis")
	}

	if err = session.DeleteRefreshTokenByUserID(ctx, client, userID); err != nil {
		return errors.ErrInternalServerError.Wrap(err, "failed to delete refresh token")
	}

	return nil

}
