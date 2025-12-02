package api

import (
	"github.com/gin-gonic/gin"

	"evm_event_indexer/api/controller/v1/contracts"
	"evm_event_indexer/api/middleware"
)

func Routing(router *gin.Engine) {

	v1 := router.Group("/v1")
	{

		// user := v1.Group("/user")
		// {
		// 	user.POST("/create", middleware.AuthValidation(), users.Create)
		// 	// user.GET("/list", middleware.AuthValidation(), userController.List)
		// 	// user.GET("/get/:id", middleware.AuthValidation(), userController.Get)
		// }

		log := v1.Group("/txn")
		{
			// Add more routes here as needed
			log.GET("/logs", middleware.AuthValidation(), contracts.GetLog)
			// get block
			// get transaction
			// get receipt
			// get event log
			// get event detail
		}

	}
}
