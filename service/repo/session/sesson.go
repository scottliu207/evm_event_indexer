package session

import (
	"context"
	"encoding/json"
	"evm_event_indexer/internal/config"
	"evm_event_indexer/internal/errors"
	"evm_event_indexer/internal/session"
	"evm_event_indexer/internal/storage"
	"evm_event_indexer/service/model"
	"evm_event_indexer/utils/hashing"
	"fmt"

	"github.com/redis/go-redis/v9"
)

// newUserRTKey is the key for the user refresh token, value is session_id
func newUserRTKey(rt string) string {
	return fmt.Sprintf("user_rt:%s", rt)
}

// newUserCSRFKey is the key for the user csrf token, value is session_id
func newUserCSRFKey(csrf string) string {
	return fmt.Sprintf("user_csrf:%s", csrf)
}

// newSessionKey is the key for the session, value is session_data
func newSessionKey(sessionID string) string {
	return fmt.Sprintf("session:%s", sessionID)
}

// newUserSessionKey is the key for the user session, value is session_id
func newUserSessionKey(userID int64) string {
	return fmt.Sprintf("user_session:%d", userID)
}

// CreateSession creates a session for a user, if the user has an old session, it will also be revoked.
func CreateSession(ctx context.Context, client *redis.Client, userID int64) (*model.SessionOut, error) {

	if userID == 0 {
		return nil, errors.ErrApiInvalidParam.New("userID is empty")
	}

	oldSessionID, err := client.Get(ctx, newUserSessionKey(userID)).Result()
	if err != nil && err != redis.Nil {
		return nil, fmt.Errorf("failed to get session data, error: %w", err)
	}

	pipe := client.TxPipeline()

	// if old session exists, revoke all related sessions
	if oldSessionID != "" {

		// get old session data
		oldData, err := GetSessionData(ctx, client, oldSessionID)
		if err != nil {
			return nil, fmt.Errorf("failed to get session data, error: %w", err)
		}

		if oldData != nil {
			// revoke old RT
			pipe.Del(ctx, newUserRTKey(oldData.HashedRT))
			// revoke old session
			pipe.Del(ctx, newSessionKey(oldData.ID))
			// revoke old user session
			pipe.Del(ctx, newUserSessionKey(oldData.UserID))
			// revoke old csrf token
			pipe.Del(ctx, newUserCSRFKey(oldData.HashedCSRF))
		}
	}

	// create session id
	sessionID, err := session.NewSessionID()
	if err != nil {
		return nil, err
	}

	// create access token
	at, tokenObj, err := session.NewJWT(config.Get().Session.JWTSecret).GenerateToken(userID, sessionID, config.Get().Session.ATExpiration)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token, err: %w", err)
	}

	claims, ok := tokenObj.Claims.(*session.JwtClaim)
	if !ok {
		return nil, fmt.Errorf("failed to assert token claims")
	}

	// create refresh token
	plainRT, hashedRT, err := session.NewOpaqueToken()
	if err != nil {
		return nil, fmt.Errorf("failed to generate opaque token, err: %w", err)
	}

	// create csrf token
	plainCSRF, hashedCSRF, err := session.NewOpaqueToken()
	if err != nil {
		return nil, fmt.Errorf("failed to generate opaque token, err: %w", err)
	}

	res := model.Session{
		ID:     sessionID,
		UserID: userID,
		AT:     at,
	}

	newStore := &model.SessionStore{
		Session:    res,
		HashedRT:   hashedRT,
		HashedCSRF: hashedCSRF,
	}

	// store session id by refresh token
	pipe.Set(ctx, newUserRTKey(hashedRT), sessionID, config.Get().Session.SessionExpiration)
	// store session data
	pipe.Set(ctx, newSessionKey(sessionID), newStore, config.Get().Session.SessionExpiration)
	// store session id by user id
	pipe.Set(ctx, newUserSessionKey(userID), sessionID, config.Get().Session.SessionExpiration)
	// store session id by csrf token
	pipe.Set(ctx, newUserCSRFKey(hashedCSRF), sessionID, config.Get().Session.SessionExpiration)

	_, err = pipe.Exec(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create refresh token: %w", err)
	}

	return &model.SessionOut{
		Session:     res,
		RT:          plainRT,
		CSRFToken:   plainCSRF,
		ATExpiresAt: claims.ExpiresAt.Time,
	}, nil
}

// VerifyAccessToken verifies a user access token
// return 0 and nil if the access token is invalid
func VerifyAccessToken(ctx context.Context, at string) (*session.JwtClaim, error) {
	claims, err := session.NewJWT(config.Get().Session.JWTSecret).VerifyToken(at)
	if err != nil {
		return nil, err
	}

	if claims == nil {
		return nil, nil
	}

	client, err := storage.GetRedis(config.RedisCert)
	if err != nil {
		return nil, fmt.Errorf("failed to get redis: %w", err)
	}

	data, err := GetSessionData(ctx, client, claims.Sid)
	if err != nil {
		return nil, fmt.Errorf("failed to get session data: %w", err)
	}

	if data == nil {
		return nil, nil
	}

	return claims, nil
}

// RevokeUserSessionByUserID revoke a user session by user id
func RevokeUserSessionByUserID(ctx context.Context, client *redis.Client, userID int64) error {

	// get session id by user id
	sessionID, err := client.Get(ctx, newUserSessionKey(userID)).Result()
	if err != nil && err != redis.Nil {
		return fmt.Errorf("failed to get user session: %w", err)
	}

	// if session id is not found, skip revoke
	if sessionID == "" {
		return nil
	}

	// get session data by session id
	dataStr, err := client.Get(ctx, newSessionKey(sessionID)).Result()
	if err != nil && err != redis.Nil {
		return fmt.Errorf("failed to get session data: %w", err)
	}

	pipe := client.TxPipeline()
	pipe.Del(ctx, newUserSessionKey(userID))

	if dataStr != "" {
		data := new(model.SessionStore)
		if err := json.Unmarshal([]byte(dataStr), data); err != nil {
			return fmt.Errorf("failed to unmarshal session data")
		}

		pipe.Del(ctx, newSessionKey(data.ID))
		pipe.Del(ctx, newUserRTKey(data.HashedRT))
	}
	_, err = pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete refresh token: %w", err)
	}

	return nil
}

// GetSessionIDByRT gets a session id by refresh token
func GetSessionIDByRT(ctx context.Context, client *redis.Client, rt string) (string, error) {
	sessionID, err := client.Get(ctx, newUserRTKey(hashing.Sha256([]byte(rt)))).Result()
	if err != nil && err != redis.Nil {
		return "", fmt.Errorf("failed to get session id: %w", err)
	}

	return sessionID, nil
}

// GetSessionIDByCSRF gets a session id by csrf token
func GetSessionIDByCSRF(ctx context.Context, client *redis.Client, hashedCSRF string) (string, error) {
	sessionID, err := client.Get(ctx, newUserCSRFKey(hashedCSRF)).Result()
	if err != nil && err != redis.Nil {
		return "", fmt.Errorf("failed to get session id: %w", err)
	}
	return sessionID, nil
}

// GetSessionData gets session data by session id
func GetSessionData(ctx context.Context, client *redis.Client, sessionID string) (*model.SessionStore, error) {
	dataStr, err := client.Get(ctx, newSessionKey(sessionID)).Result()
	if err != nil && err != redis.Nil {
		return nil, fmt.Errorf("failed to get session data: %w", err)
	}

	if dataStr == "" {
		return nil, nil
	}

	data := new(model.SessionStore)
	if err := json.Unmarshal([]byte(dataStr), data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal session data")
	}
	return data, nil
}
