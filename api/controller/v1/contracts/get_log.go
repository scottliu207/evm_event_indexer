package contracts

import (
	"encoding/hex"
	"evm_event_indexer/api/middleware"
	"evm_event_indexer/internal/errors"
	"evm_event_indexer/service"
	"evm_event_indexer/service/model"
	logRepo "evm_event_indexer/service/repo/eventlog"
	"net/http"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/gin-gonic/gin"
)

type (
	GetLogReq struct {
		ChainID   int64    `form:"chain_id" binding:"required"`
		Topics    []string `form:"topics" binding:"required,dive,hex"`
		Address   string   `form:"address" binding:"required"`
		StartTime string   `form:"start_time" binding:"required"`
		EndTime   string   `form:"end_time" binding:"required"`
		Page      int32    `form:"page" binding:"required,min=1"`
		Size      int32    `form:"size" binding:"required,min=1,max=100"`
		OrderBy   int8     `form:"order_by"`
		Desc      bool     `form:"desc"`
	}

	GetLogRes struct {
		Logs  []*EventLog `json:"logs"`
		Total int64       `json:"total"`
	}

	EventLog struct {
		ID              int64               `json:"id"`
		ChainID         int64               `json:"chain_id"`
		BlockNumber     uint64              `json:"block_number"`
		BlockHash       string              `json:"block_hash"`
		TransactionHash string              `json:"transaction_hash"`
		Address         string              `json:"address"`
		Topics          []common.Hash       `json:"topics"`
		Data            string              `json:"data"`
		TxIndex         int32               `json:"tx_index"`
		LogIndex        int32               `json:"log_index"`
		DecodedEvent    *model.DecodedEvent `json:"decoded_event"`
		BlockTimestamp  time.Time           `json:"block_timestamp"`
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
	if !common.IsHexAddress(req.Address) {
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

	param := &logRepo.GetLogParam{
		ChainID:   req.ChainID,
		Address:   req.Address,
		StartTime: startTime,
		EndTime:   endTime,
		OrderBy:   req.OrderBy,
		Desc:      req.Desc,
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
		res.Logs[i] = &EventLog{
			ID:              log.ID,
			ChainID:         log.ChainID,
			BlockNumber:     log.BlockNumber,
			BlockHash:       log.BlockHash,
			TransactionHash: log.TxHash,
			Address:         log.Address,
			Topics:          log.Topics.Array(),
			Data:            "0x" + hex.EncodeToString(log.Data),
			LogIndex:        log.LogIndex,
			TxIndex:         log.TxIndex,
			DecodedEvent:    log.DecodedEvent,
			BlockTimestamp:  log.BlockTimestamp,
		}
	}

	c.Status(http.StatusOK)
}
