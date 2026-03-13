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

// Delete removes a user by ID.
//
//	@Summary		Delete user
//	@Description	Delete a user by ID (admin only).
//	@Tags			Admin Users
//	@Produce		json
//	@Param			user_id	path	int	true	"User ID"
//	@Success		204		"No Content"
//	@Failure		401		{object}	protocol.Response
//	@Failure		404		{object}	protocol.Response
//	@Security		AdminBearerAuth
//	@Router			/v1/admin/users/{user_id} [delete]
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
