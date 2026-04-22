package config

import (
	"os"
	"path"
	"path/filepath"
	"runtime"
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
	HTTPServer        *HTTPServer
	APIKey            string
	StreamTokenExpiry time.Duration
	ReqlogBinaryPath  string
}

type HTTPServer struct {
	URL             string
	ShutdownTimeout time.Duration
}

func NewConfig(log logger.Logger) (*Config, error) {
	return NewConfigWithOptions(LoaderOptions{
		EnvPath: path.Join(rootDir(), "..", ".env"),
		Logger:  log,
	})
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
		log.Info("Loaded environment variables from" + opts.EnvPath)
	} else {
		log.Info("failed to load .env file, using system environment variables")
	}

	cfg := &Config{
		HTTPServer: &HTTPServer{
			URL:             getEnv("HTTP_SERVER_URL", ":4000"),
			ShutdownTimeout: getEnvDuration("HTTP_SERVER_SHUTDOWN_TIMEOUT", 5*time.Second),
		},
		APIKey:            getEnv("HTTP_AUTH_API_KEY", ""),
		ReqlogBinaryPath:  getEnv("REQLOG_BINARY_PATH", "reqlog"),
		StreamTokenExpiry: getEnvDuration("STREAM_TOKEN_EXPIRY", 30*time.Second),
	}

	return cfg, nil
}

func rootDir() string {
	_, b, _, _ := runtime.Caller(0)
	d := path.Join(path.Dir(b))

	return filepath.Dir(d)
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
