package server

import (
	"context"
	"fmt"
	"io/fs"
	"net"
	"net/http"
	"slices"

	"github.com/gin-gonic/gin"
	"github.com/sagarmaheshwary/reqlog-ui/internal/config"
	"github.com/sagarmaheshwary/reqlog-ui/internal/logger"
	"github.com/sagarmaheshwary/reqlog-ui/internal/service"
	"github.com/sagarmaheshwary/reqlog-ui/internal/tokenstore"
	"github.com/sagarmaheshwary/reqlog-ui/internal/transports/http/server/handler"
	"github.com/sagarmaheshwary/reqlog-ui/internal/transports/http/server/middleware"
	"github.com/sagarmaheshwary/reqlog-ui/internal/web"
)

type Opts struct {
	Config        *config.HTTPServer
	Logger        logger.Logger
	ReqlogService service.ReqlogService
	TokenStore    *tokenstore.Store
	ReqlogConfig  *config.Reqlog
}

type HTTPServer struct {
	config *config.HTTPServer
	server *http.Server
	logger logger.Logger
}

func NewServer(opts *Opts) *HTTPServer {
	ginMode := opts.Config.GinMode
	if !slices.Contains([]string{gin.DebugMode, gin.ReleaseMode}, ginMode) {
		panic(fmt.Errorf("invalid gin mode: %s", ginMode))
	}

	gin.SetMode(ginMode)
	r := gin.New()

	r.Use(
		gin.Recovery(),
	)

	if opts.Config.Logger {
		r.Use(middleware.Logger(opts.Logger))
	}

	sub, err := fs.Sub(web.StaticFS, "static")
	if err != nil {
		panic(err)
	}

	r.StaticFS("/static", http.FS(sub))

	r.GET("/", serveHTML(sub, "index.html"))
	r.GET("/login", serveHTML(sub, "login.html"))

	api := r.Group("/api")

	authHandler := handler.NewAuthHandler(&handler.AuthHandlerOpts{
		APIKey:     opts.Config.APIKey,
		TokenStore: opts.TokenStore,
	})
	api.POST("/auth/token", authHandler.Token)

	protected := api.Group("/")
	protected.Use(middleware.APIKeyAuth(opts.Config.APIKey))
	{
		// Issues an expirable single-use token for the SSE endpoint
		protected.POST("/auth/stream-token", authHandler.StreamToken)

		reqlogHandler := handler.NewReqlogHandler(&handler.ReqlogHandlerOpts{
			ReqlogService: opts.ReqlogService,
			Logger:        opts.Logger,
			Config:        opts.ReqlogConfig,
		})

		protected.GET("/logs", reqlogHandler.Logs)
		// SSE uses its own token-based auth so the API key never hits a URL
		api.GET("/logs/stream",
			middleware.StreamTokenAuth(opts.TokenStore),
			reqlogHandler.LogsStream,
		)
	}

	return &HTTPServer{
		config: opts.Config,
		server: &http.Server{
			Addr:    opts.Config.URL,
			Handler: r,
		},
		logger: opts.Logger,
	}
}

func (h *HTTPServer) ServeListener(listener net.Listener) error {
	h.logger.Info("HTTP server started", logger.Field{Key: "address", Value: listener.Addr().String()})
	if err := h.server.Serve(listener); err != nil && err != http.ErrServerClosed {
		h.logger.Error("HTTP server failed", logger.Field{Key: "error", Value: err.Error()})
		return err
	}
	return nil
}

func (h *HTTPServer) Serve() error {
	listener, err := net.Listen("tcp", h.config.URL)
	if err != nil {
		h.logger.Error("Failed to create HTTP listener",
			logger.Field{Key: "address", Value: h.config.URL},
			logger.Field{Key: "error", Value: err.Error()},
		)
		return err
	}

	return h.ServeListener(listener)
}

func (h *HTTPServer) Shutdown(ctx context.Context) error {
	if h.server == nil {
		return nil
	}
	return h.server.Shutdown(ctx)
}

func serveHTML(sub fs.FS, file string) func(c *gin.Context) {
	return func(c *gin.Context) {
		data, err := fs.ReadFile(sub, file)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			return
		}
		c.Data(200, "text/html; charset=utf-8", data)
	}
}
