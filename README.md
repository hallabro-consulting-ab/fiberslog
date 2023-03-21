# fiberslog

fiberslog is a [golang.org/x/exp/slog](https://pkg.go.dev/golang.org/x/exp/slog) logging middleware for [Fiber](https://github.com/gofiber/fiber), a web framework for Go.
This middleware logs incoming HTTP requests and outgoing HTTP responses, including their respective status code, timestamp and latency.

## Installation

You can install the package via Go modules:

```sh
go get github.com/hallabro-consulting-ab/fiberslog
```

## Usage

```go
import "github.com/hallabro-consulting-ab/fiberslog"

app := fiber.New()
app.Use(fiberslog.New(options...))
```

Where options is a list of `fiberslog.Option` functions. The following options are available:

- `fiberslog.WithLogger(logger *slog.Logger)` sets the logger to use.
- `fiberslog.WithPopulateContext(populate bool)` sets whether to populate the request context with the logger. With this option set you can access the logger by calling ```logger := ctx.Locals("logger")```.
- `fiberslog.WithNext(next func(ctx *fiber.Ctx) bool)` sets the next handler to call. This is useful if you want to skip the middleware for whatever reason.
