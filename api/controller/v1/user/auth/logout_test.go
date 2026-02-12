package auth_test

import (
	"bytes"
	"encoding/json"
	"evm_event_indexer/api/controller/v1/user/auth"
	"evm_event_indexer/api/protocol"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLogout_Success(t *testing.T) {
	reqBody := auth.LogoutReq{}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var res protocol.Response
	err := json.Unmarshal(w.Body.Bytes(), &res)
	assert.NoError(t, err)
	assert.Equal(t, 0, res.Code)
	assert.Equal(t, "success", res.Message)
	assert.Empty(t, res.Result)
}
