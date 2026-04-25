package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"runtime/debug"

	"github.com/sagarmaheshwary/reqlog-ui/internal/config"
	"github.com/sagarmaheshwary/reqlog-ui/internal/logger"
	"github.com/sagarmaheshwary/reqlog-ui/internal/service"
	"github.com/sagarmaheshwary/reqlog-ui/internal/tokenstore"
	"github.com/sagarmaheshwary/reqlog-ui/internal/transports/http/server"
)

var (
	version     = "dev"
	showVersion = flag.Bool("version", false, "print version and exit")
)

func main() {
	flag.Parse()

	if *showVersion {
		fmt.Printf("reqlog-ui version %s\n", getVersion())
		return
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	log := logger.NewZerologLogger("info", os.Stderr)

	cfg, err := config.NewConfig(log)
	if err != nil {
		log.Fatal(err.Error())
	}

	reqlogService := service.NewReqlogService(service.ReqlogServiceOpts{
		Config: cfg.Reqlog,
	})
	tokenStore := tokenstore.New(ctx, cfg.HTTPServer.StreamTokenExpiry)

	httpServer := server.NewServer(&server.Opts{
		Config:        cfg.HTTPServer,
		Logger:        log,
		ReqlogService: reqlogService,
		TokenStore:    tokenStore,
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

func getVersion() string {
	if version != "dev" {
		return version
	}

	info, ok := debug.ReadBuildInfo()
	if !ok {
		return "dev"
	}

	if info.Main.Version != "" && info.Main.Version != "(devel)" {
		return info.Main.Version
	}

	return "dev"
}
