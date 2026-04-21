package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"

	"github.com/sagarmaheshwary/reqlog-ui/internal/config"
	"github.com/sagarmaheshwary/reqlog-ui/internal/logger"
	"github.com/sagarmaheshwary/reqlog-ui/internal/transports/http/server"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	log := logger.NewZerologLogger("info", os.Stderr)

	cfg, err := config.NewConfig(log)
	if err != nil {
		log.Fatal(err.Error())
	}

	httpServer := server.NewServer(&server.Opts{
		Config: cfg.HTTPServer,
		Logger: log,
	})
	go func() {
		err = httpServer.Serve()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			stop()
		}
	}()

	<-ctx.Done()

	log.Warn("Shutdown signal received, closing services!")

	httpCtx, httpCancel := context.WithTimeout(context.Background(), cfg.HTTPServer.ShutdownTimeout)
	if err := httpServer.Server.Shutdown(httpCtx); err != nil {
		log.Error("failed to close http server", logger.Field{Key: "error", Value: err.Error()})
	}
	httpCancel()

	log.Info("Shutdown complete!")
}
