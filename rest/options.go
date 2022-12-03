package rest

import (
	"net/http"

	"github.com/go-chi/cors"
	"go.uber.org/zap"
)

const (
	DefaultGracefulShutdownSec = 5

	DefaultPort = "8080"
)

type config struct {
	port        string
	middlewares []func(next http.Handler) http.Handler

	allowedHosts []string

	setCors     bool
	corsOptions cors.Options

	logger *zap.Logger
}

type Option func(*config) error

func WithPort(port string) Option {
	return func(c *config) error {
		c.port = port
		return nil
	}
}

func WithMiddlewares(middlewares []func(next http.Handler) http.Handler) Option {
	return func(c *config) error {
		c.middlewares = middlewares
		return nil
	}
}

func WithAllowedHosts(allowedHosts []string) Option {
	return func(c *config) error {
		c.allowedHosts = allowedHosts
		return nil
	}
}

func WithCoreOptions(corsOptions cors.Options) Option {
	return func(c *config) error {
		c.corsOptions = corsOptions
		c.setCors = true
		return nil
	}
}

func WithZapLogger(logger *zap.Logger) Option {
	return func(c *config) error {
		c.logger = logger
		return nil
	}
}

