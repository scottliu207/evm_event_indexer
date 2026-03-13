package api

import (
	"evm_event_indexer/api/middleware"
	internalCnf "evm_event_indexer/internal/config"
	"net/http"

	_ "evm_event_indexer/docs" // swagger generated docs

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// NewServer creates a new HTTP server
func NewServer() *http.Server {

	cnf := internalCnf.Get()
	// gin.SetMode(gin.ReleaseMode)
	router := gin.New()

	// register logger
	router.Use(middleware.Logger())

	router.Use(middleware.CORS(middleware.Options{
		AllowCredentials: true,
	}))

	// register swagger UI before timeout/response middleware to avoid Content-Length conflict
	if cnf.API.EnableSwagger {
		router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}

	// handle unexpected request
	router.NoMethod(middleware.NotFoundHandler)
	router.NoRoute(middleware.NotFoundHandler)

	// use recovery middleware to avoid panic
	router.Use(gin.Recovery(), middleware.Interceptor(), middleware.TimeoutHandler(), middleware.ResponseHandler())

	// bind route
	Routing(router)

	listenPort := ":" + cnf.API.Port

	// http server
	server := &http.Server{
		Addr:    listenPort,
		Handler: router,
	}

	return server
}
