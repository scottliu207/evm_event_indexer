package contracts_test

import (
	"context"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"evm_event_indexer/api/protocol"
	"evm_event_indexer/internal/config"
	"evm_event_indexer/internal/session"
	"evm_event_indexer/internal/storage"
	"evm_event_indexer/service/model"
	logRepo "evm_event_indexer/service/repo/eventlog"
	"evm_event_indexer/utils"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetLog_Success(t *testing.T) {
	db, err := storage.GetMySQL(config.EventDBM)
	require.NoError(t, err)

	chainID := int64(31337)
	address := common.HexToAddress(fmt.Sprintf("0x%040x", time.Now().UnixNano())).Hex()

	transferSig := "0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef"
	fromAddr := common.HexToAddress("0x0000000000000000000000000000000000000000")
	toAddr := common.HexToAddress("0xf39fd6e51aad88f6f4ce6ab8827279cfffb92266")
	fromTopic := common.BytesToHash(common.LeftPadBytes(fromAddr.Bytes(), 32)).Hex()
	toTopic := common.BytesToHash(common.LeftPadBytes(toAddr.Bytes(), 32)).Hex()

	blockTimestamp := time.Now()
	logRow := &model.Log{
		ChainID:        chainID,
		Address:        address,
		BlockHash:      common.HexToHash("0x01").Hex(),
		BlockNumber:    1,
		Topic0:         transferSig,
		Topic1:         fromTopic,
		Topic2:         toTopic,
		Topic3:         "",
		TxIndex:        0,
		LogIndex:       0,
		TxHash:         common.HexToHash("0x02").Hex(),
		Data:           common.LeftPadBytes(common.Big1.Bytes(), 32),
		DecodedEvent:   &model.DecodedEvent{EventName: "Transfer", EventData: map[string]string{"from": fromAddr.Hex(), "to": toAddr.Hex(), "value": "1"}},
		BlockTimestamp: blockTimestamp,
		CreatedAt:      time.Now(),
	}

	require.NoError(t, utils.NewTx(db).Exec(ctx, func(ctx context.Context, tx *sql.Tx) error {
		return logRepo.TxInsertLog(ctx, tx, logRow)
	}))
	t.Cleanup(func() {
		_ = utils.NewTx(db).Exec(ctx, func(ctx context.Context, tx *sql.Tx) error {
			return logRepo.TxDeleteLog(ctx, tx, address, 0)
		})
	})

	start := blockTimestamp.Add(-24 * time.Hour).UTC().Format(time.RFC3339)
	end := blockTimestamp.Add(24 * time.Hour).UTC().Format(time.RFC3339)

	at, err := session.NewJWT(config.Get().Session.JWTSecret).GenerateToken(1, time.Hour)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet,
		fmt.Sprintf("/api/v1/txn/logs?chain_id=%d&address=%s&signature=%s&from=%s&to=%s&start_time=%s&end_time=%s&page=1&size=10",
			chainID,
			address,
			transferSig,
			fromAddr.Hex(),
			toAddr.Hex(),
			start,
			end,
		),
		nil,
	)
	req.Header.Set("Authorization", "Bearer "+at)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	var res protocol.Response
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &res))
	assert.Equal(t, 0, res.Code)
	assert.Equal(t, "success", res.Message)

	result, ok := res.Result.(map[string]any)
	assert.True(t, ok)

	total, ok := result["total"].(float64)
	assert.True(t, ok)
	assert.Equal(t, float64(1), total)

	logs, ok := result["logs"].([]any)
	assert.True(t, ok)
	assert.Len(t, logs, 1)

	first, ok := logs[0].(map[string]any)
	assert.True(t, ok)

	assert.Equal(t, float64(chainID), first["chain_id"])
	assert.Equal(t, address, first["address"])
	assert.Equal(t, float64(1), first["block_number"])
	assert.Equal(t, transferSig, first["topics"].([]any)[0])

	dataHex, ok := first["data"].(string)
	assert.True(t, ok)
	assert.True(t, len(dataHex) > 2 && dataHex[:2] == "0x")
	_, err = hex.DecodeString(dataHex[2:])
	assert.NoError(t, err)

	decoded, ok := first["decoded_event"].(map[string]any)
	assert.True(t, ok)
	assert.Equal(t, "Transfer", decoded["event_name"])
}

func TestGetLog_InvalidToken(t *testing.T) {
	// test invalid token
	testCases := []struct {
		name         string
		token        string
		expectedCode int
	}{
		{
			name:         "missing access token",
			token:        "",
			expectedCode: http.StatusUnauthorized,
		},
		{
			name:         "invalid access token",
			token:        "invalid_token",
			expectedCode: http.StatusUnauthorized,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			req := httptest.NewRequest(http.MethodGet, "/api/v1/txn/logs", nil)
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+tc.token)
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

func TestGetLog_InvalidRequest(t *testing.T) {
	db, err := storage.GetMySQL(config.EventDBM)
	assert.NoError(t, err)

	chainID := int64(31337)
	address := common.HexToAddress(fmt.Sprintf("0x%040x", time.Now().UnixNano())).Hex()

	transferSig := "0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef"
	fromAddr := common.HexToAddress("0x0000000000000000000000000000000000000000")
	toAddr := common.HexToAddress("0xf39fd6e51aad88f6f4ce6ab8827279cfffb92266")

	fromTopic := common.BytesToHash(common.LeftPadBytes(fromAddr.Bytes(), 32)).Hex()
	toTopic := common.BytesToHash(common.LeftPadBytes(toAddr.Bytes(), 32)).Hex()

	blockTimestamp := time.Now()
	logRow := &model.Log{
		ChainID:        chainID,
		Address:        address,
		BlockHash:      common.HexToHash("0x01").Hex(),
		BlockNumber:    1,
		Topic0:         transferSig,
		Topic1:         fromTopic,
		Topic2:         toTopic,
		Topic3:         "",
		TxIndex:        0,
		LogIndex:       0,
		TxHash:         common.HexToHash("0x02").Hex(),
		Data:           common.LeftPadBytes(common.Big1.Bytes(), 32),
		DecodedEvent:   &model.DecodedEvent{EventName: "Transfer", EventData: map[string]string{"from": fromAddr.Hex(), "to": toAddr.Hex(), "value": "1"}},
		BlockTimestamp: blockTimestamp,
		CreatedAt:      time.Now(),
	}

	assert.NoError(t, utils.NewTx(db).Exec(ctx, func(ctx context.Context, tx *sql.Tx) error {
		return logRepo.TxInsertLog(ctx, tx, logRow)
	}))
	t.Cleanup(func() {
		_ = utils.NewTx(db).Exec(ctx, func(ctx context.Context, tx *sql.Tx) error {
			return logRepo.TxDeleteLog(ctx, tx, address, 0)
		})
	})

	start := blockTimestamp.Add(-24 * time.Hour).UTC().Format(time.RFC3339)
	end := blockTimestamp.Add(24 * time.Hour).UTC().Format(time.RFC3339)

	at, err := session.NewJWT(config.Get().Session.JWTSecret).GenerateToken(1, time.Hour)
	assert.NoError(t, err)

	// test invalid request
	testCases := []struct {
		name         string
		params       map[string]string
		expectedCode int
	}{
		{
			name:         "missing start_time",
			params:       map[string]string{"chain_id": strconv.FormatInt(chainID, 10), "address": address, "signature": transferSig, "from": fromAddr.Hex(), "to": toAddr.Hex(), "end_time": end, "page": "1", "size": "10"},
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "missing end_time",
			params:       map[string]string{"chain_id": strconv.FormatInt(chainID, 10), "address": address, "signature": transferSig, "from": fromAddr.Hex(), "to": toAddr.Hex(), "start_time": start, "page": "1", "size": "10"},
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "missing page",
			params:       map[string]string{"chain_id": strconv.FormatInt(chainID, 10), "address": address, "signature": transferSig, "from": fromAddr.Hex(), "to": toAddr.Hex(), "start_time": start, "end_time": end, "size": "10"},
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "missing size",
			params:       map[string]string{"chain_id": strconv.FormatInt(chainID, 10), "address": address, "signature": transferSig, "from": fromAddr.Hex(), "to": toAddr.Hex(), "start_time": start, "end_time": end, "page": "1"},
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "invalid page",
			params:       map[string]string{"chain_id": strconv.FormatInt(chainID, 10), "address": address, "signature": transferSig, "from": fromAddr.Hex(), "to": toAddr.Hex(), "start_time": start, "end_time": end, "page": "0", "size": "10"},
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "invalid size",
			params:       map[string]string{"chain_id": strconv.FormatInt(chainID, 10), "address": address, "signature": transferSig, "from": fromAddr.Hex(), "to": toAddr.Hex(), "start_time": start, "end_time": end, "page": "1", "size": "0"},
			expectedCode: http.StatusBadRequest,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			req := httptest.NewRequest(http.MethodGet,
				fmt.Sprintf("/api/v1/txn/logs?chain_id=%s&address=%s&signature=%s&from=%s&to=%s&start_time=%s&end_time=%s&page=%s&size=%s",
					tc.params["chain_id"],
					tc.params["address"],
					tc.params["signature"],
					tc.params["from"],
					tc.params["to"],
					tc.params["start_time"],
					tc.params["end_time"],
					tc.params["page"],
					tc.params["size"],
				), nil)

			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+at)
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
