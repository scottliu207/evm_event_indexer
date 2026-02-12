package users

import (
	"net/http"

	"evm_event_indexer/service"

	"github.com/gin-gonic/gin"
)

type (
	DeleteReq struct {
		UserID int64 `uri:"user_id" binding:"required,min=1"`
	}
)

func Delete(c *gin.Context) {

	var req = new(DeleteReq)
	if err := c.ShouldBindUri(req); err != nil {
		c.Error(err)
		return
	}

	if err := service.DeleteUserByAdmin(c.Request.Context(), req.UserID); err != nil {
		c.Error(err)
		return
	}

	c.Status(http.StatusNoContent)
}
