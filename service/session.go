package service

import (
	"context"
	"evm_event_indexer/internal/config"
	"evm_event_indexer/internal/errors"
	"evm_event_indexer/internal/storage"
	"evm_event_indexer/service/model"
	"evm_event_indexer/service/repo/session"
	"evm_event_indexer/utils/hashing"
	"strconv"
)

// VerifyCSRFToken verifies a csrf token
func VerifyCSRFToken(ctx context.Context, csrf string) (*model.SessionStore, error) {
	if csrf == "" {
		return nil, errors.ErrCSRFTokenInvalid.New("csrf token is required")
	}

	client, err := storage.GetRedis(config.RedisCert)
	if err != nil {
		return nil, errors.ErrInternalServerError.Wrap(err, "failed to get redis")
	}

	hashed := hashing.Sha256([]byte(csrf))

	// get session_id by csrf token
	sessionID, err := session.GetSessionIDByCSRF(ctx, client, hashed)
	if err != nil {
		return nil, errors.ErrInternalServerError.Wrap(err, "failed to get session id by refresh token")
	}

	if sessionID == "" {
		return nil, errors.ErrInvalidCredentials.New("invalid csrf token")
	}

	// get session data by session_id
	data, err := session.GetSessionData(ctx, client, sessionID)
	if err != nil {
		return nil, errors.ErrInternalServerError.Wrap(err, "failed to get session id by refresh token")
	}

	if data == nil || data.HashedCSRF != hashed {
		return nil, errors.ErrCSRFTokenInvalid.New("invalid csrf token")
	}

	return data, nil
}

// VerifyUserAT verifies a user access token
func VerifyUserAT(ctx context.Context, at string) (int64, error) {
	claims, err := session.VerifyAccessToken(ctx, at)
	if err != nil {
		return 0, errors.ErrInternalServerError.Wrap(err)
	}

	if claims == nil {
		return 0, errors.ErrInvalidCredentials.New("invalid access token")
	}

	userID, err := strconv.ParseInt(claims.Subject, 10, 64)
	if err != nil {
		return 0, errors.ErrInternalServerError.Wrap(err, "failed to parse user id")
	}

	return userID, nil
}

// GetUserByRT gets a userID by refresh token, return InvalidCredentialsError if the refresh token is invalid.
func GetUserIDByRT(ctx context.Context, rt string) (int64, error) {
	client, err := storage.GetRedis(config.RedisCert)
	if err != nil {
		return 0, errors.ErrInternalServerError.Wrap(err, "failed to get redis")
	}

	// get session id by refresh token
	sessionID, err := session.GetSessionIDByRT(ctx, client, rt)
	if err != nil {
		return 0, errors.ErrInternalServerError.Wrap(err, "failed to get user id by refresh token")
	}

	if sessionID == "" {
		return 0, errors.ErrInvalidCredentials.New("invalid refresh token")
	}

	data, err := session.GetSessionData(ctx, client, sessionID)
	if err != nil {
		return 0, errors.ErrInternalServerError.Wrap(err, "failed to get session data")
	}

	// if session data is not found, return InvalidCredentialsError
	if data == nil {
		return 0, errors.ErrInvalidCredentials.New("invalid refresh token")
	}

	return data.UserID, nil
}

// DeleteUserRT deletes a user refresh token
func RevokeUserSession(ctx context.Context, userID int64) error {

	if userID == 0 {
		return errors.ErrApiInvalidParam.New("user id is 0")
	}

	client, err := storage.GetRedis(config.RedisCert)
	if err != nil {
		return errors.ErrInternalServerError.Wrap(err, "failed to get redis")
	}

	if err = session.RevokeUserSessionByUserID(ctx, client, userID); err != nil {
		return errors.ErrInternalServerError.Wrap(err, "failed to revoke user session")
	}

	return nil
}

// CreateSession creates a session for a user
func CreateSession(ctx context.Context, userID int64) (*model.SessionOut, error) {

	if userID == 0 {
		return nil, errors.ErrApiInvalidParam.New("userID")
	}

	client, err := storage.GetRedis(config.RedisCert)
	if err != nil {
		return nil, errors.ErrInternalServerError.Wrap(err, "failed to get redis")
	}

	session, err := session.CreateSession(ctx, client, userID)
	if err != nil {
		return nil, errors.ErrInternalServerError.Wrap(err, "failed to create session")
	}

	return session, nil
}
