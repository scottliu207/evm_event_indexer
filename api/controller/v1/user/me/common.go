package me

import (
	"evm_event_indexer/api/middleware"
	"evm_event_indexer/internal/errors"

	"github.com/gin-gonic/gin"
)

func getMyID(c *gin.Context) (int64, error) {
	userID := c.GetInt64(middleware.CtxUserID)
	if userID <= 0 {
		return 0, errors.ErrInvalidCredentials.New("user id not found")
	}
	return userID, nil
}
