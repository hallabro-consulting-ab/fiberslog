package fiberslog

import (
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/exp/slog"
)

type Config struct {
	Next            func(ctx *fiber.Ctx) bool
	Logger          *slog.Logger
	PopulateContext bool
}

type Option func(*Config)

func WithNext(next func(ctx *fiber.Ctx) bool) Option {
	return func(c *Config) {
		c.Next = next
	}
}

func WithLogger(logger *slog.Logger) Option {
	return func(c *Config) {
		c.Logger = logger
	}
}

func WithPopulateContext(populate bool) Option {
	return func(c *Config) {
		c.PopulateContext = populate
	}
}

func New(options ...Option) fiber.Handler {
	var (
		cfg    *Config
		logger *slog.Logger
	)

	cfg = new(Config)

	if len(options) > 0 {
		for _, option := range options {
			if option != nil {
				option(cfg)
			}
		}
	}

	if cfg.Logger == nil {
		logger = slog.Default()
	} else {
		logger = cfg.Logger
	}

	return func(c *fiber.Ctx) error {
		if cfg.PopulateContext {
			c.Locals("logger", logger)
		}

		if cfg.Next != nil && cfg.Next(c) {
			return c.Next()
		}

		start := time.Now()

		msg := "incoming request"
		err := c.Next()
		if err != nil {
			msg = err.Error()
		}

		code := c.Response().StatusCode()

		dumplogger := logger.With(
			slog.Int("status", code),
			slog.String("method", c.Method()),
			slog.String("path", c.Path()),
			slog.String("ip", c.IP()),
			slog.String("latency", time.Since(start).String()),
			slog.String("user-agent", c.Get(fiber.HeaderUserAgent)))

		switch {
		case code >= fiber.StatusBadRequest && code < fiber.StatusInternalServerError:
			dumplogger.Warn(msg)
		case code >= http.StatusInternalServerError:
			dumplogger.Error(msg)
		default:
			dumplogger.Info(msg)
		}

		return err
	}
}
