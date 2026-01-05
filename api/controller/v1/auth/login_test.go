package auth_test

import (
	"bytes"
	"encoding/json"
	"evm_event_indexer/api/controller/v1/auth"
	"evm_event_indexer/api/protocol"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestLogin_Success tests successful login
func TestLogin_Success(t *testing.T) {
	reqBody := auth.LoginReq{
		Account:  testAccount,
		Password: testPassword,
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var res protocol.Response
	err := json.Unmarshal(w.Body.Bytes(), &res)
	assert.NoError(t, err)

	assert.Equal(t, 0, res.Code)
	assert.Equal(t, "success", res.Message)

	// Check access token exists in response
	result, ok := res.Result.(map[string]any)
	assert.True(t, ok, "result should be a map")
	assert.NotEmpty(t, result["access_token"], "access_token should not be empty")
	assert.NotEmpty(t, result["csrf_token"], "csrf_token should not be empty")

	// Check refresh token cookie is set
	cookies := w.Result().Cookies()
	var refreshTokenCookie *http.Cookie
	var csrfTokenCookie *http.Cookie
	for _, c := range cookies {
		if c.Name == "refresh_token" {
			refreshTokenCookie = c
		}
		if c.Name == "csrf_token" {
			csrfTokenCookie = c
		}
	}
	assert.NotNil(t, refreshTokenCookie, "refresh_token cookie should be set")
	assert.True(t, refreshTokenCookie.HttpOnly, "refresh_token cookie should be HttpOnly")
	assert.True(t, refreshTokenCookie.Secure, "refresh_token cookie should be Secure")
	assert.NotNil(t, csrfTokenCookie, "csrf_token cookie should be set")
	assert.False(t, csrfTokenCookie.HttpOnly, "csrf_token cookie should not be HttpOnly")
	assert.True(t, csrfTokenCookie.Secure, "csrf_token cookie should be Secure")
}

// TestLogin_InvalidCredentials tests login with wrong password
func TestLogin_InvalidCredentials(t *testing.T) {
	testCases := []struct {
		name     string
		account  string
		password string
	}{
		{
			name:     "wrong password",
			account:  testAccount,
			password: "wrong_password",
		},
		{
			name:     "non-existent account",
			account:  "non_existent_user",
			password: testPassword,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			reqBody := auth.LoginReq{
				Account:  tc.account,
				Password: tc.password,
			}
			body, _ := json.Marshal(reqBody)

			req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusUnauthorized, w.Code)

			var res protocol.Response
			err := json.Unmarshal(w.Body.Bytes(), &res)
			assert.NoError(t, err)
			assert.NotEqual(t, 0, res.Code, "error code should not be 0")
		})
	}
}

// TestLogin_InvalidRequest tests login with invalid request body
func TestLogin_InvalidRequest(t *testing.T) {
	testCases := []struct {
		name         string
		body         any
		expectedCode int
	}{
		{
			name:         "missing account",
			body:         map[string]string{"password": testPassword},
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "missing password",
			body:         map[string]string{"account": testAccount},
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "empty body",
			body:         map[string]string{},
			expectedCode: http.StatusBadRequest,
		},
		{
			name: "account too long",
			// LoginReq.Account has binding max=50
			body:         map[string]string{"account": "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", "password": testPassword}, // 51 chars
			expectedCode: http.StatusBadRequest,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			body, _ := json.Marshal(tc.body)

			req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tc.expectedCode, w.Code)

			var res protocol.Response
			err := json.Unmarshal(w.Body.Bytes(), &res)
			assert.NoError(t, err)
			assert.NotEqual(t, 0, res.Code, "error code should not be 0")
		})
	}
}

// TestLogin_InvalidJSON tests login with malformed JSON
func TestLogin_InvalidJSON(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBufferString(`{"account": invalid}`))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
