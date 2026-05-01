package shutdown

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"time"

	"file_service/pkg/logging"
)

func Gracful(signals []os.Signal, server *http.Server) {
	logger := logging.GetLogger()
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, signals...)
	sig := <-stop
	logger.Warn("shutdown signal received", "signal", sig.String())

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		logger.Error("graceful shutdown failed", "error", err)
		return
	}

	logger.Info("server shutdown completed")
}
