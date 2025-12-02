package contracts

import (
	"evm_event_indexer/api/middleware"
	"evm_event_indexer/internal/errors"
	"evm_event_indexer/service/model"
	logRepo "evm_event_indexer/service/repo/eventlog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type (
	GetLogReq struct {
		Address   string    `form:"address" binding:"required"`
		StartTime time.Time `form:"start_time" binding:"required"`
		EndTime   time.Time `form:"end_time" binding:"required"`
		Page      int32     `form:"page" binding:"required,min=1"`
		Size      int32     `form:"size" binding:"required,min=1,max=100"`
	}

	GetLogRes struct {
		Logs  []*EventLog `json:"logs"`
		Total int64       `json:"total"`
	}

	EventLog struct {
		ID              int64     `json:"id"`
		BlockNumber     uint64    `json:"block_number"`
		BlockHash       string    `json:"block_hash"`
		TransactionHash string    `json:"transaction_hash"`
		Address         string    `json:"address"`
		Topics          []string  `json:"topics"`
		Data            []byte    `json:"data"`
		TxIndex         int32     `json:"tx_index"`
		LogIndex        int32     `json:"log_index"`
		BlockTimestamp  time.Time `json:"block_timestamp"`
	}
)

func GetLog(c *gin.Context) {
	res := new(GetLogRes)
	res.Logs = make([]*EventLog, 0)

	c.Set(middleware.CTX_RESPONSE, res)

	var req GetLogReq
	if err := c.ShouldBindQuery(&req); err != nil {
		c.Error(err)
		return
	}

	param := &logRepo.GetLogParam{
		Address:   req.Address,
		StartTime: req.StartTime,
		EndTime:   req.EndTime,
		Pagination: &model.Pagination{
			Page: req.Page,
			Size: req.Size,
		},
	}

	total, err := logRepo.GetTotal(c.Request.Context(), param)
	if err != nil {
		c.Error(errors.INTERNAL_SERVER_ERROR.Wrap(err, "failed to get event log total"))
		return
	}

	// no need to proceed if no data found
	if total == 0 {
		c.Status(http.StatusOK)
		return
	}

	res.Total = total

	logs, err := logRepo.GetLogs(c.Request.Context(), param)
	if err != nil {
		c.Error(errors.INTERNAL_SERVER_ERROR.Wrap(err, "failed to get event logs"))
		return
	}

	res.Logs = make([]*EventLog, len(logs))
	for i, log := range logs {
		res.Logs[i] = &EventLog{
			ID:              log.ID,
			BlockNumber:     log.BlockNumber,
			BlockHash:       log.BlockHash,
			TransactionHash: log.TxHash,
			Address:         log.Address,
			Topics:          log.Topics.Array(),
			Data:            log.Data,
			LogIndex:        log.LogIndex,
			TxIndex:         log.TxIndex,
			BlockTimestamp:  log.BlockTimestamp,
		}
	}

	c.Status(http.StatusOK)
}
