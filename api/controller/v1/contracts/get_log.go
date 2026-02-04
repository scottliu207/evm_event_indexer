package contracts

import (
	"encoding/hex"
	"evm_event_indexer/api/middleware"
	"evm_event_indexer/internal/errors"
	"evm_event_indexer/service"
	"evm_event_indexer/service/model"
	logRepo "evm_event_indexer/service/repo/eventlog"
	"net/http"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/gin-gonic/gin"
)

type (
	GetLogReq struct {
		ChainID   int64  `form:"chain_id"`
		Address   string `form:"address" binding:"omitempty"`
		TxHash    string `form:"tx_hash" binding:"omitempty"`
		BNStart   uint64 `form:"bn_start" binding:"omitempty"` // 0 means no limit
		BNEnd     uint64 `form:"bn_end" binding:"omitempty"`   // 0 means no limit
		Signature string `form:"signature" binding:"omitempty"`
		From      string `form:"from" binding:"omitempty"`
		To        string `form:"to" binding:"omitempty"`
		StartTime string `form:"start_time" binding:"required"`
		EndTime   string `form:"end_time" binding:"required"`
		OrderBy   int8   `form:"order_by"`
		Desc      bool   `form:"desc"`
		Page      uint64 `form:"page" binding:"required,min=1"`
		Size      uint64 `form:"size" binding:"required,min=1,max=100"`
	}

	GetLogRes struct {
		Logs  []*EventLog `json:"logs"`
		Total int64       `json:"total"`
	}

	EventLog struct {
		ID             int64               `json:"id"`
		ChainID        int64               `json:"chain_id"`
		BlockNumber    uint64              `json:"block_number"`
		BlockHash      string              `json:"block_hash"`
		TxHash         string              `json:"tx_hash"`
		Address        string              `json:"address"`
		Topics         []string            `json:"topics"`
		Data           string              `json:"data"`
		TxIndex        int32               `json:"tx_index"`
		LogIndex       int32               `json:"log_index"`
		DecodedEvent   *model.DecodedEvent `json:"decoded_event"`
		BlockTimestamp time.Time           `json:"block_timestamp"`
	}
)

// GetLog retrieves event logs
func GetLog(c *gin.Context) {
	res := new(GetLogRes)
	res.Logs = make([]*EventLog, 0)

	c.Set(middleware.CtxResponse, res)

	var req GetLogReq
	if err := c.ShouldBindQuery(&req); err != nil {
		c.Error(err)
		return
	}

	// Validate address format
	if req.Address != "" && !common.IsHexAddress(req.Address) {
		c.Error(errors.ErrApiInvalidParam.Wrap(nil, "invalid address format"))
		return
	}

	// Parse start_time (RFC3339 format)
	startTime, err := time.Parse(time.RFC3339, req.StartTime)
	if err != nil {
		c.Error(errors.ErrApiInvalidParam.Wrap(err, "invalid start_time format, expected RFC3339"))
		return
	}

	// Parse end_time (RFC3339 format)
	endTime, err := time.Parse(time.RFC3339, req.EndTime)
	if err != nil {
		c.Error(errors.ErrApiInvalidParam.Wrap(err, "invalid end_time format, expected RFC3339"))
		return
	}

	// default order by block number
	if req.OrderBy == 0 {
		req.OrderBy = 2
	}

	if req.Signature != "" && !isHex32Bytes(req.Signature) {
		c.Error(errors.ErrApiInvalidParam.New("invalid signature, expected 32-byte hex"))
		return
	}

	fromTopic, err := normalizeTopicAddressOrHash(req.From)
	if err != nil {
		c.Error(err)
		return
	}
	toTopic, err := normalizeTopicAddressOrHash(req.To)
	if err != nil {
		c.Error(err)
		return
	}

	param := &logRepo.GetLogParam{
		ChainID:        req.ChainID,
		Address:        req.Address,
		StartTime:      startTime,
		EndTime:        endTime,
		Topic0:         strings.ToLower(strings.TrimSpace(req.Signature)),
		Topic1:         fromTopic,
		Topic2:         toTopic,
		TxHash:         req.TxHash,
		BlockNumberGTE: req.BNStart,
		BlockNumberLTE: req.BNEnd,
		OrderBy:        req.OrderBy,
		Desc:           req.Desc,
		Pagination: &model.Pagination{
			Page: req.Page,
			Size: req.Size,
		},
	}

	logs, total, err := service.GetLogsWithTotal(c.Request.Context(), param)
	if err != nil {
		c.Error(err)
		return
	}

	res.Total = total
	res.Logs = make([]*EventLog, len(logs))
	for i, log := range logs {

		topics := make([]string, 0)
		if log.Topic0 != "" {
			topics = append(topics, log.Topic0)
		}
		if log.Topic1 != "" {
			topics = append(topics, log.Topic1)
		}
		if log.Topic2 != "" {
			topics = append(topics, log.Topic2)
		}
		if log.Topic3 != "" {
			topics = append(topics, log.Topic3)
		}

		res.Logs[i] = &EventLog{
			ID:             log.ID,
			ChainID:        log.ChainID,
			BlockNumber:    log.BlockNumber,
			BlockHash:      log.BlockHash,
			TxHash:         log.TxHash,
			Address:        log.Address,
			Topics:         topics,
			Data:           "0x" + hex.EncodeToString(log.Data),
			LogIndex:       log.LogIndex,
			TxIndex:        log.TxIndex,
			DecodedEvent:   log.DecodedEvent,
			BlockTimestamp: log.BlockTimestamp,
		}
	}

	c.Status(http.StatusOK)
}

func isHex32Bytes(s string) bool {
	s = strings.TrimSpace(s)
	if len(s) != 66 || !strings.HasPrefix(s, "0x") {
		return false
	}
	_, err := hex.DecodeString(s[2:])
	return err == nil
}

func normalizeTopicAddressOrHash(s string) (string, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return "", nil
	}

	if common.IsHexAddress(s) {
		addr := common.HexToAddress(s)
		padded := common.LeftPadBytes(addr.Bytes(), 32)
		return common.BytesToHash(padded).Hex(), nil
	}

	if isHex32Bytes(s) {
		return strings.ToLower(s), nil
	}

	return "", errors.ErrApiInvalidParam.New("invalid topic format, expected address or 32-byte hex")
}
