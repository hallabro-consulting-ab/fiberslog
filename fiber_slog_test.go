package fiberslog_test

import (
	"errors"
	"github.com/hallabro-consulting-ab/fiberslog"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/exp/slog"
)

func TestLog(t *testing.T) {
	for _, tc := range []struct {
		name            string
		statusCode      int
		handlerErr      error
		expectedLogLine *regexp.Regexp
		options         []fiberslog.Option
	}{
		{
			name:       "success",
			statusCode: http.StatusOK,
			handlerErr: nil,
			// time=2023-03-12T17:56:01.823+01:00 level=INFO msg=\"incoming request\" status=200 method=GET path=/ ip=0.0.0.0 latency=5.209µs user-agent=test
			expectedLogLine: regexp.MustCompile(`time=\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}.\d{3}(\+\d{2}:\d{2})?Z? level=INFO msg="incoming request" status=200 method=GET path=/ ip=0.0.0.0 latency=\d+(\.\d+)?\S+ user-agent=test\n`),
		},
		{
			name:       "error",
			statusCode: http.StatusInternalServerError,
			handlerErr: errors.New("test error"),
			// time=2023-03-12T17:56:01.823+01:00 level=ERROR msg=\"test error\" status=500 method=GET path=/ ip=0.0.0.0 latency=5.209µs user-agent=test
			expectedLogLine: regexp.MustCompile(`time=\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}.\d{3}(\+\d{2}:\d{2})?Z? level=ERROR msg="test error" status=500 method=GET path=/ ip=0.0.0.0 latency=\d+(\.\d+)?\S+ user-agent=test\n`),
		},
		{
			name:       "warning",
			statusCode: http.StatusBadRequest,
			handlerErr: errors.New("invalid data"),
			// time=2023-03-12T17:56:01.823+01:00 level=WARN msg=\"test error\" status=500 method=GET path=/ ip=0.0.0.0 latency=5.209µs user-agent=test
			expectedLogLine: regexp.MustCompile(`time=\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}.\d{3}(\+\d{2}:\d{2})?Z? level=WARN msg="invalid data" status=400 method=GET path=/ ip=0.0.0.0 latency=\d+(\.\d+)?\S+ user-agent=test\n`),
		},
		{
			name:       "warning",
			statusCode: http.StatusBadRequest,
			handlerErr: errors.New("invalid data"),
			// time=2023-03-12T17:56:01.823+01:00 level=WARN msg=\"test error\" status=500 method=GET path=/ ip=0.0.0.0 latency=5.209µs user-agent=test
			expectedLogLine: regexp.MustCompile(`time=\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}.\d{3}(\+\d{2}:\d{2})?Z? level=WARN msg="invalid data" status=400 method=GET path=/ ip=0.0.0.0 latency=\d+(\.\d+)?\S+ user-agent=test\n`),
		},
		{
			name: "populate context",
			options: []fiberslog.Option{
				fiberslog.WithPopulateContext(true),
				fiberslog.WithNext(func(ctx *fiber.Ctx) bool {
					logger := ctx.Locals("logger")
					assert.NotNil(t, logger)
					if logger == nil {
						return false
					}

					l, ok := logger.(*slog.Logger)
					assert.True(t, ok)
					if !ok {
						return false
					}

					l.Info("from handler")

					return false
				}),
			},
			statusCode:      http.StatusOK,
			expectedLogLine: regexp.MustCompile(`^time=\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}.\d{3}(\+\d{2}:\d{2})?Z? level=INFO msg="from handler"*`),
		},
		{
			name: "do not populate context",
			options: []fiberslog.Option{
				fiberslog.WithPopulateContext(false),
				fiberslog.WithNext(func(ctx *fiber.Ctx) bool {
					logger := ctx.Locals("logger")
					assert.Nil(t, logger)

					return false
				}),
			},
			statusCode: http.StatusOK,
		},
	} {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			logs := &strings.Builder{}

			options := []fiberslog.Option{
				fiberslog.WithLogger(slog.New(slog.NewTextHandler(logs))),
				fiberslog.WithNext(
					func(ctx *fiber.Ctx) bool {
						return false
					},
				),
			}

			if tc.options != nil && len(tc.options) > 0 {
				options = append(options, tc.options...)
			}

			slogMiddleware := fiberslog.New(options...)

			app := fiber.New()
			app.Use(slogMiddleware)

			app.Get("/", func(c *fiber.Ctx) error {
				err := c.SendStatus(tc.statusCode)
				require.NoError(t, err)

				return tc.handlerErr
			})

			req := httptest.NewRequest("GET", "/", nil)
			req.Header.Set("User-Agent", "test")

			_, err := app.Test(req)
			require.NoError(t, err)

			if tc.expectedLogLine != nil {
				assert.Regexp(t, tc.expectedLogLine, logs.String())
			}
		})
	}
}
