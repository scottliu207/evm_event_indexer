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
	router.Use(middleware.ResponseHandler(), middleware.TimeoutHandler(), gin.Recovery())

	// bind route
	Routing(router)

	listenPort := ":" + cnf.API.Port

	// start http server
	server := &http.Server{
		Addr:    listenPort,
		Handler: router,
	}

	err := server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		slog.Error("API server error.", slog.Any("error", err))
		os.Exit(1)
	}

	slog.Info("API server is running.", slog.Any("port", cnf.API.Port))
}
