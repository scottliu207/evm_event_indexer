package api

import (
	"net/http"

	adminAuthController "evm_event_indexer/api/controller/v1/admin/auth"
	adminUsersController "evm_event_indexer/api/controller/v1/admin/users"

	"github.com/gin-gonic/gin"

	"evm_event_indexer/api/controller/v1/contracts"
	authController "evm_event_indexer/api/controller/v1/user/auth"
	"evm_event_indexer/api/controller/v1/user/me"
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
				auth.POST("/logout", middleware.Authorization(), authController.Logout)
				auth.POST("/refresh", middleware.CSRFProtection(), authController.RotateToken)
			}

			user := v1.Group("/user")
			{

				managed := user.Group("", middleware.Authorization())
				{
					managed.GET("/me", me.GetMe)
					managed.PUT("/me", me.UpdateMe)
				}
			}

			admin := v1.Group("/admin")
			{
				adminAuth := admin.Group("/auth")
				{
					adminAuth.POST("/login", adminAuthController.Login)
					adminAuth.POST("/logout", middleware.AdminAuthorization(), adminAuthController.Logout)
					adminAuth.POST("/refresh", middleware.AdminCSRFProtection(), adminAuthController.RotateToken)
				}

				adminUsers := admin.Group("/users", middleware.AdminAuthorization())
				{
					adminUsers.POST("", adminUsersController.Create)
					adminUsers.GET("", adminUsersController.List)
					adminUsers.GET("/:user_id", adminUsersController.Get)
					adminUsers.PUT("/:user_id", adminUsersController.Update)
					adminUsers.DELETE("/:user_id", adminUsersController.Delete)
				}
			}

			log := v1.Group("/txn", middleware.Authorization())
			{
				// Add more routes here as needed
				log.GET("/logs", contracts.GetLog)
				// get block
				// get transaction
				// get receipt
				// get event log
				// get event detail
			}

		}
	}
}
