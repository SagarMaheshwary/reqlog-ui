package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"runtime/debug"

	"github.com/rs/zerolog"
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

	var out io.Writer
	out = zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: "2006-01-02T15:04:05.000Z"}
	if os.Getenv("DISABLE_PRETTY_LOGS") == "1" {
		out = os.Stderr
	}

	log := logger.NewZerologLogger("info", out)

	envPath := os.Getenv("ENV_FILE")
	if envPath == "" {
		envPath = ".env"
	}

	cfg, err := config.NewConfigWithOptions(config.LoaderOptions{
		EnvPath: envPath,
		Logger:  log,
	})
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
