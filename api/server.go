package api

import (
	"evm_event_indexer/api/middleware"
	internalCnf "evm_event_indexer/internal/config"
	"log/slog"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

// Start RESTful API Service
func Listen() {

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

	// start http server
	server := &http.Server{
		Addr:    listenPort,
		Handler: router,
	}

	slog.Info("API server is running.", slog.Any("port", cnf.API.Port))
	err := server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		slog.Error("failed to start API server", slog.Any("error", err))
		os.Exit(1)
	}

}
