package api

import (
	"github.com/gin-gonic/gin"

	userController "evm_event_indexer/api/controller/v1/user"
	"evm_event_indexer/api/middleware"
)

func Routing(router *gin.Engine) {

	// deploy contract
	// get block
	// get transaction
	// get receipt
	// get event log
	// get event detail

	v1 := router.Group("/v1")
	{

		user := v1.Group("/user")
		{
			user.POST("/create", middleware.AuthValidation(), userController.Create)
			// user.GET("/list", middleware.AuthValidation(), userController.List)
			// user.GET("/get/:id", middleware.AuthValidation(), userController.Get)
		}

		// log := v1.Group("/log")
		{
			// Add more routes here as needed

		}

	}
}
