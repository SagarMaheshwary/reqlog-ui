package config

import (
	"os"
	"strconv"
	"time"

	"github.com/gofor-little/env"
	"github.com/sagarmaheshwary/reqlog-ui/internal/logger"
)

type LoaderOptions struct {
	EnvPath   string
	EnvLoader func(string) error
	Logger    logger.Logger
}

type Config struct {
	HTTPServer *HTTPServer
	Reqlog     *Reqlog
}

type Reqlog struct {
	BinaryPath       string
	ExecutionTimeout time.Duration
}

type HTTPServer struct {
	URL               string
	ShutdownTimeout   time.Duration
	GinMode           string
	Logger            bool
	APIKey            string
	StreamTokenExpiry time.Duration
}

func NewConfigWithOptions(opts LoaderOptions) (*Config, error) {
	log := opts.Logger
	envLoader := opts.EnvLoader
	if envLoader == nil {
		envLoader = func(path string) error {
			_, err := os.Stat(path)
			if err != nil {
				return err
			}

			return env.Load(path)
		}
	}

	if err := envLoader(opts.EnvPath); err == nil {
		log.Info("Loaded environment variables from " + opts.EnvPath)
	} else {
		log.Info("failed to load .env file, using system environment variables", logger.Field{Key: "path", Value: opts.EnvPath})
	}

	cfg := &Config{
		HTTPServer: &HTTPServer{
			URL:               getEnv("HTTP_SERVER_URL", "localhost:4000"),
			ShutdownTimeout:   getEnvDuration("HTTP_SERVER_SHUTDOWN_TIMEOUT", 5*time.Second),
			GinMode:           getEnv("HTTP_GIN_MODE", "release"),
			Logger:            getEnvBool("HTTP_LOGGER_ENABLED", false),
			APIKey:            getEnv("HTTP_AUTH_API_KEY", ""),
			StreamTokenExpiry: getEnvDuration("HTTP_STREAM_TOKEN_EXPIRY", 30*time.Second),
		},
		Reqlog: &Reqlog{
			BinaryPath:       getEnv("REQLOG_BINARY_PATH", "reqlog"),
			ExecutionTimeout: getEnvDuration("REQLOG_EXECUTION_TIMEOUT", 15*time.Minute),
		},
	}

	return cfg, nil
}

func getEnv(key string, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}

	return defaultVal
}

func getEnvDuration(key string, defaultVal time.Duration) time.Duration {
	if val, err := time.ParseDuration(os.Getenv(key)); err == nil {
		return val
	}

	return defaultVal
}

func getEnvBool(key string, defaultVal bool) bool {
	if val, err := strconv.ParseBool(os.Getenv(key)); err == nil {
		return val
	}

	return defaultVal
}
