package middleware

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/goccy/go-json"
	"github.com/labstack/echo/v4"
	"golang.org/x/exp/slog"
)

const requestIDCtx = "slog-echo.request-id"

type Config struct {
	DefaultLevel     slog.Level
	ClientErrorLevel slog.Level
	ServerErrorLevel slog.Level

	WithRequestID bool
}

// NewLogger returns a echo.MiddlewareFunc (middleware) that logs requests using slog.
//
// Requests with errors are logged using slog.Error().
// Requests without errors are logged using slog.Info().
func NewLogger(logger *slog.Logger) echo.MiddlewareFunc {
	return NewLoggerWithConfig(logger, Config{
		DefaultLevel:     slog.LevelInfo,
		ClientErrorLevel: slog.LevelWarn,
		ServerErrorLevel: slog.LevelError,

		WithRequestID: true,
	})
}

// NewLoggerWithConfig returns a echo.HandlerFunc (middleware) that logs requests using slog.
func NewLoggerWithConfig(logger *slog.Logger, config Config) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) (err error) {
			req := c.Request()
			res := c.Response()

			start := time.Now()

			path := c.Path()
			if path == "" {
				path = req.URL.Path
			}

			respWrt := &responseWriter{
				ResponseWriter: c.Response().Writer,
			}

			c.Response().Writer = respWrt

			reqBody := getRequestBody(c)

			err = next(c)

			requestID := req.Header.Get(echo.HeaderXRequestID)
			if requestID == "" {
				requestID = res.Header().Get(echo.HeaderXRequestID)
			}

			status := res.Status
			method := req.Method
			end := time.Now()
			latency := end.Sub(start)
			ip := c.RealIP()
			userAgent := req.UserAgent()

			httpErr := new(echo.HTTPError)
			if err != nil && errors.As(err, &httpErr) {
				status = httpErr.Code
				if msg, ok := httpErr.Message.(string); ok {
					err = errors.New(msg)
				}
			}

			attributes := []slog.Attr{
				slog.Time("time", end),
				slog.String("path", path),
				slog.String("remote-ip", ip),
				slog.String("user-agent", userAgent),
				slog.Group("request",
					slog.String("method", method),
					slog.String("uri", req.URL.Path),
					slog.Any("body", reqBody),
				),
				slog.Group("response",
					slog.Any("body", bytesToMap(respWrt.body)),
					slog.String("latency", latency.String()),
					slog.Int("status", status),
				),
			}

			if err != nil {
				attributes = append(attributes, slog.String("error", err.Error()))
			}

			if config.WithRequestID {
				attributes = append(attributes, slog.String("request-id", requestID))
			}

			msg := fmt.Sprintf("%d %s", res.Status, req.URL.Path)

			switch {
			case status >= http.StatusInternalServerError:
				logger.LogAttrs(context.Background(), config.ServerErrorLevel, msg, attributes...)
			case status >= http.StatusBadRequest && status < http.StatusInternalServerError:
				logger.LogAttrs(context.Background(), config.ClientErrorLevel, msg, attributes...)
			case status >= http.StatusMultipleChoices && status < http.StatusBadRequest:
				attributes = append(attributes, slog.Bool("redirect", true))
				logger.LogAttrs(context.Background(), config.DefaultLevel, msg, attributes...)
			default:
				logger.LogAttrs(context.Background(), config.DefaultLevel, msg, attributes...)
			}

			return
		}
	}
}

// GetRequestID returns the request identifier
func GetRequestID(c echo.Context) string {
	if id, ok := c.Get(requestIDCtx).(string); ok {
		return id
	}

	return ""
}

type responseWriter struct {
	http.ResponseWriter
	body []byte
	code int
}

func (w *responseWriter) WriteHeader(code int) {
	w.code = code
	w.ResponseWriter.WriteHeader(code)
}

func (w *responseWriter) Write(b []byte) (int, error) {
	w.body = cloneBytes(b)
	return w.ResponseWriter.Write(b)
}

func cloneBytes(b []byte) []byte {
	c := make([]byte, len(b))
	copy(c, b)
	return c
}

func getRequestBody(c echo.Context) any {
	if c.Request().Method == echo.GET || c.Request().Body == nil {
		return nil
	}

	bs, _ := io.ReadAll(c.Request().Body)
	c.Request().Body = io.NopCloser(bytes.NewReader(bs))
	c.Request().Body.Close()

	if len(bs) == 0 {
		return nil
	}

	return bytesToMap(bs)
}

func bytesToMap(bs []byte) any {

	if len(bs) == 0 {
		return nil
	}

	// obj
	if bs[0] == '{' {
		m := make(map[string]any)
		json.Unmarshal(bs, &m)
		return m
	}

	// array
	if bs[0] == '[' {
		var arr []map[string]any
		json.Unmarshal(bs, &arr)
		return arr
	}

	return bs
}
