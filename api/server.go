package api

import (
	"evm_event_indexer/api/middleware"
	internalCnf "evm_event_indexer/internal/config"
	"net/http"

	_ "evm_event_indexer/docs" // swagger generated docs

	"github.com/gin-gonic/gin"
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
