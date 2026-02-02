package session

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_JWT(t *testing.T) {
	session := NewJWT("secret")
	sessionID := "session_1"
	token, _, err := session.GenerateToken(1, sessionID, 2*time.Second)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)
	t.Logf("token: %s", token)

	// test valid token
	claims, err := session.VerifyToken(token)
	assert.NoError(t, err)
	if assert.NotNil(t, claims) {
		assert.Equal(t, "1", claims.Subject)
		assert.Equal(t, sessionID, claims.Sid)
	}

	// test invalid token
	claims, err = session.VerifyToken("invalid_token")
	assert.NoError(t, err)
	assert.Nil(t, claims)

	// test valid token with bearer prefix
	claims, err = session.VerifyToken("bearer " + token)
	assert.NoError(t, err)
	if assert.NotNil(t, claims) {
		assert.Equal(t, "1", claims.Subject)
		assert.Equal(t, sessionID, claims.Sid)
	}

	// test valid token with Bearer prefix (canonical form)
	claims, err = session.VerifyToken("Bearer " + token)
	assert.NoError(t, err)
	if assert.NotNil(t, claims) {
		assert.Equal(t, "1", claims.Subject)
		assert.Equal(t, sessionID, claims.Sid)
	}

	// wait for token to expire
	time.Sleep(3 * time.Second)
	claims, err = session.VerifyToken(token)
	assert.NoError(t, err)
	assert.Nil(t, claims)
}
