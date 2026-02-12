package background

import (
	"context"
	"evm_event_indexer/internal/config"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type MetricsServer struct {
	server *http.Server
}

func NewMetricsServer() *MetricsServer {

	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	return &MetricsServer{
		server: &http.Server{
			Addr:    ":" + config.Get().Metrics.Port,
			Handler: mux,
		},
	}
}

func (s *MetricsServer) Run(ctx context.Context) error {

	go func() {
		slog.Info("metrics server is running", slog.String("port", config.Get().Metrics.Port))
		if err := s.server.ListenAndServe(); err != nil {
			slog.Error("metrics server listen and serve error", slog.Any("error", err))
		}
	}()

	<-ctx.Done()

	// at this point, parent context has been cancelled
	// create a new context with timeout for shutting down the server gracefully
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := s.server.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("shutdown metrics server error: %w", err)
	}

	slog.Info("metrics server is stopped")
	return nil
}
