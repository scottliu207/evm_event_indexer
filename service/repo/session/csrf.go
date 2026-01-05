package session

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"evm_event_indexer/internal/config"
	"evm_event_indexer/internal/errors"
)

// NewCSRFToken creates a new CSRF token
func NewCSRFToken(token string) (string, error) {
	secret := config.Get().Session.CSRFSecret
	if secret == "" {
		return "", errors.ErrInternalServerError.New("csrf secret is empty")
	}

	if token == "" {
		return "", errors.ErrInternalServerError.New("refresh token is empty")
	}

	h := hmac.New(sha256.New, []byte(secret))
	_, _ = h.Write([]byte(token))
	sum := h.Sum(nil)
	return base64.RawURLEncoding.EncodeToString(sum), nil
}
