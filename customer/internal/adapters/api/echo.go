package api

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	gommonLog "github.com/labstack/gommon/log"
	ctxutil "github.com/vterry/food-project/common/pkg/context"
	"github.com/vterry/food-project/common/pkg/errors"
)

const CorrelationIDHeader = "X-Correlation-ID"

// NewEcho creates a new Echo instance with middlewares and configuration local to the Customer service.
func NewEcho() *echo.Echo {
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true

	e.HTTPErrorHandler = CustomHTTPErrorHandler
	e.Logger = NewEchoSlogAdapter(slog.Default())

	e.Use(middleware.Recover())
	e.Use(middleware.ContextTimeout(10 * time.Second))

	// Correlation ID Middleware
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			req := c.Request()
			correlationID := req.Header.Get(CorrelationIDHeader)
			ctx := ctxutil.WithCorrelationID(req.Context(), correlationID)
			correlationID = ctxutil.GetCorrelationID(ctx)
			c.SetRequest(req.WithContext(ctx))
			c.Response().Header().Set(CorrelationIDHeader, correlationID)
			return next(c)
		}
	})

	// Request Logger Middleware
	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogStatus:   true,
		LogURI:      true,
		LogMethod:   true,
		LogLatency:  true,
		LogError:    true,
		HandleError: true,
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			ctx := c.Request().Context()
			correlationID := ctxutil.GetCorrelationID(ctx)

			if v.Error != nil {
				slog.ErrorContext(ctx, "Request processed with error",
					slog.String("method", v.Method),
					slog.String("uri", v.URI),
					slog.Int("status", v.Status),
					slog.Duration("latency", v.Latency),
					slog.String("err", v.Error.Error()),
					slog.String("correlation_id", correlationID),
				)
			} else {
				slog.InfoContext(ctx, "Request processed",
					slog.String("method", v.Method),
					slog.String("uri", v.URI),
					slog.Int("status", v.Status),
					slog.Duration("latency", v.Latency),
					slog.String("correlation_id", correlationID),
				)
			}
			return nil
		},
	}))

	e.GET("/health", func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	})

	return e
}

// CustomHTTPErrorHandler handles errors for the Echo framework using common mapping logic.
func CustomHTTPErrorHandler(err error, c echo.Context) {
	if c.Response().Committed {
		return
	}

	response, statusCode := errors.MapToHTTP(err)

	slog.ErrorContext(c.Request().Context(), "Handling HTTP error",
		slog.String("err", err.Error()),
		slog.Int("status", statusCode),
		slog.String("slug", response.Slug),
	)

	if err := c.JSON(statusCode, response); err != nil {
		slog.ErrorContext(c.Request().Context(), "Failed to send error response", slog.String("err", err.Error()))
	}
}

// EchoSlogAdapter adapts slog to Echo's Logger interface.
type EchoSlogAdapter struct {
	logger *slog.Logger
	level  gommonLog.Lvl
	prefix string
	output io.Writer
}

func NewEchoSlogAdapter(logger *slog.Logger) *EchoSlogAdapter {
	return &EchoSlogAdapter{
		logger: logger,
		level:  gommonLog.INFO,
		output: os.Stdout,
	}
}

func (e *EchoSlogAdapter) jsonToAttrs(j gommonLog.JSON) []slog.Attr {
	attrs := make([]slog.Attr, 0, len(j))
	for k, v := range j {
		attrs = append(attrs, slog.Any(k, v))
	}
	return attrs
}

func (e *EchoSlogAdapter) Output() io.Writer            { return e.output }
func (e *EchoSlogAdapter) SetOutput(w io.Writer)        { e.output = w }
func (e *EchoSlogAdapter) Prefix() string               { return e.prefix }
func (e *EchoSlogAdapter) SetPrefix(p string)           { e.prefix = p }
func (e *EchoSlogAdapter) Level() gommonLog.Lvl         { return e.level }
func (e *EchoSlogAdapter) SetLevel(v gommonLog.Lvl)     { e.level = v }
func (e *EchoSlogAdapter) SetHeader(h string)           {}
func (e *EchoSlogAdapter) Print(i ...interface{})       { e.logger.Info(fmt.Sprint(i...)) }
func (e *EchoSlogAdapter) Printf(f string, a ...interface{}) { e.logger.Info(fmt.Sprintf(f, a...)) }
func (e *EchoSlogAdapter) Printj(j gommonLog.JSON) {
	e.logger.LogAttrs(context.Background(), slog.LevelInfo, "echo", e.jsonToAttrs(j)...)
}
func (e *EchoSlogAdapter) Debug(i ...interface{})       { e.logger.Debug(fmt.Sprint(i...)) }
func (e *EchoSlogAdapter) Debugf(f string, a ...interface{}) { e.logger.Debug(fmt.Sprintf(f, a...)) }
func (e *EchoSlogAdapter) Debugj(j gommonLog.JSON) {
	e.logger.LogAttrs(context.Background(), slog.LevelDebug, "echo", e.jsonToAttrs(j)...)
}
func (e *EchoSlogAdapter) Info(i ...interface{})       { e.logger.Info(fmt.Sprint(i...)) }
func (e *EchoSlogAdapter) Infof(f string, a ...interface{}) { e.logger.Info(fmt.Sprintf(f, a...)) }
func (e *EchoSlogAdapter) Infoj(j gommonLog.JSON) {
	e.logger.LogAttrs(context.Background(), slog.LevelInfo, "echo", e.jsonToAttrs(j)...)
}
func (e *EchoSlogAdapter) Warn(i ...interface{})       { e.logger.Warn(fmt.Sprint(i...)) }
func (e *EchoSlogAdapter) Warnf(f string, a ...interface{}) { e.logger.Warn(fmt.Sprintf(f, a...)) }
func (e *EchoSlogAdapter) Warnj(j gommonLog.JSON) {
	e.logger.LogAttrs(context.Background(), slog.LevelWarn, "echo", e.jsonToAttrs(j)...)
}
func (e *EchoSlogAdapter) Error(i ...interface{})       { e.logger.Error(fmt.Sprint(i...)) }
func (e *EchoSlogAdapter) Errorf(f string, a ...interface{}) { e.logger.Error(fmt.Sprintf(f, a...)) }
func (e *EchoSlogAdapter) Errorj(j gommonLog.JSON) {
	e.logger.LogAttrs(context.Background(), slog.LevelError, "echo", e.jsonToAttrs(j)...)
}
func (e *EchoSlogAdapter) Fatal(i ...interface{}) {
	e.logger.Error(fmt.Sprint(i...))
	os.Exit(1)
}
func (e *EchoSlogAdapter) Fatalf(f string, a ...interface{}) {
	e.logger.Error(fmt.Sprintf(f, a...))
	os.Exit(1)
}
func (e *EchoSlogAdapter) Fatalj(j gommonLog.JSON) {
	e.logger.LogAttrs(context.Background(), slog.LevelError, "echo fatal", e.jsonToAttrs(j)...)
	os.Exit(1)
}
func (e *EchoSlogAdapter) Panic(i ...interface{}) {
	msg := fmt.Sprint(i...)
	e.logger.Error(msg)
	panic(msg)
}
func (e *EchoSlogAdapter) Panicf(f string, a ...interface{}) {
	msg := fmt.Sprintf(f, a...)
	e.logger.Error(msg)
	panic(msg)
}
func (e *EchoSlogAdapter) Panicj(j gommonLog.JSON) {
	e.logger.LogAttrs(context.Background(), slog.LevelError, "echo panic", e.jsonToAttrs(j)...)
	panic(j)
}
