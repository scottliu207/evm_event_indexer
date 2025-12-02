package background

import (
	"context"
	"evm_event_indexer/api"
	"evm_event_indexer/internal/config"
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

type APIServer struct {
	server *http.Server
}

func NewAPIServer() *APIServer {
	return &APIServer{
		server: api.NewServer(),
	}
}

func (s *APIServer) Run(ctx context.Context) error {

	go func() {
		slog.Info("API server is running", slog.String("port", config.Get().API.Port))
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("listen and serve error", slog.Any("err", err))
		}
	}()

	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := s.server.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("shutdown api server error: %w", err)
	}

	slog.Info("API server is stopped")
	return nil
}
