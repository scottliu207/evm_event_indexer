package api

import (
	"net/http"

	"github.com/gin-gonic/gin"

	authController "evm_event_indexer/api/controller/v1/auth"
	"evm_event_indexer/api/controller/v1/contracts"
	"evm_event_indexer/api/middleware"
)

func Routing(router *gin.Engine) {

	api := router.Group("/api")
	{

		// health check api
		api.GET("/status", func(ctx *gin.Context) {
			ctx.Status(http.StatusOK)
		})

		v1 := api.Group("/v1")
		{

			auth := v1.Group("/auth")
			{
				auth.POST("/login", authController.Login)
				auth.POST("/logout", middleware.AuthValidation(), authController.Logout)
				auth.POST("/refresh", authController.RotateToken)
			}
			// user := v1.Group("/user")
			{
				// 	user.POST("/create", middleware.AuthValidation(), users.Create)
				// 	// user.GET("/list", middleware.AuthValidation(), userController.List)
				// 	// user.GET("/get/:id", middleware.AuthValidation(), userController.Get)
			}

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
}
