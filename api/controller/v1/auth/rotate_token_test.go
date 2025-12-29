package auth_test

import (
	"bytes"
	"encoding/json"
	"evm_event_indexer/api/controller/v1/auth"
	"evm_event_indexer/api/protocol"
	"evm_event_indexer/internal/errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRotateToken_Success(t *testing.T) {
	// First login to get a real refresh token stored in Redis.
	loginBody, _ := json.Marshal(auth.LoginReq{
		Account:  testAccount,
		Password: testPassword,
	})
	loginReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBuffer(loginBody))
	loginReq.Header.Set("Content-Type", "application/json")
	loginW := httptest.NewRecorder()
	router.ServeHTTP(loginW, loginReq)
	assert.Equal(t, http.StatusOK, loginW.Code)

	var loginRes protocol.Response
	assert.NoError(t, json.Unmarshal(loginW.Body.Bytes(), &loginRes))
	assert.Equal(t, 0, loginRes.Code)

	var oldRT string
	for _, c := range loginW.Result().Cookies() {
		if c.Name == "refresh_token" {
			oldRT = c.Value
			break
		}
	}
	assert.NotEmpty(t, oldRT, "login should set refresh_token cookie")

	reqBody := auth.RotateTokenReq{}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{
		Name:     "refresh_token",
		Value:    oldRT,
		Path:     "/",
		Domain:   "localhost",
		MaxAge:   3600,
		HttpOnly: true,
		Secure:   true,
	})
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var res protocol.Response
	err := json.Unmarshal(w.Body.Bytes(), &res)
	assert.NoError(t, err)
	result, ok := res.Result.(map[string]any)
	assert.True(t, ok, "result should be a map")
	assert.Equal(t, 0, res.Code)
	assert.Equal(t, "success", res.Message)
	assert.NotEmpty(t, result["access_token"], "access_token should not be empty")

	// Check refresh token cookie is set
	cookies := w.Result().Cookies()
	var refreshTokenCookie *http.Cookie
	for _, c := range cookies {
		if c.Name == "refresh_token" {
			refreshTokenCookie = c
			break
		}
	}
	assert.NotNil(t, refreshTokenCookie, "refresh_token cookie should be set")
	assert.True(t, refreshTokenCookie.HttpOnly, "refresh_token cookie should be HttpOnly")
	assert.True(t, refreshTokenCookie.Secure, "refresh_token cookie should be Secure")
	assert.NotEqual(t, oldRT, refreshTokenCookie.Value, "refresh_token cookie should be updated")
}

func TestRotateToken_InvalidRefreshToken(t *testing.T) {
	reqBody := auth.RotateTokenReq{}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{
		Name:     "refresh_token",
		Value:    "invalid_refresh_token",
		Path:     "/",
		Domain:   "localhost",
		MaxAge:   3600,
		HttpOnly: true,
		Secure:   true,
	})
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var res protocol.Response
	err := json.Unmarshal(w.Body.Bytes(), &res)
	assert.NoError(t, err)
	assert.Equal(t, errors.ErrInvalidCredentials.ErrorCode, res.Code)
	assert.True(t, strings.HasPrefix(res.Message, errors.ErrInvalidCredentials.Message))
}
