package middleware

import (
	"time"

	"log/slog"

	"github.com/labstack/echo/v4"
)

// LoggingMiddleware logs HTTP requests with structured fields
func LoggingMiddleware(logger *slog.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()

			// Store request ID in context for later use
			traceID := c.Request().Header.Get("X-Trace-ID")
			if traceID == "" {
				traceID = c.Response().Header().Get("X-Request-ID")
			}
			c.Set("traceID", traceID)

			// Process request
			err := next(c)

			// Calculate latency
			latency := time.Since(start).Milliseconds()

			// Log with structured fields
			logger.Info("HTTP Request",
				"method", c.Request().Method,
				"path", c.Request().RequestURI,
				"status", c.Response().Status,
				"latency_ms", latency,
				"user_agent", c.Request().UserAgent(),
				"remote_ip", c.RealIP(),
				"trace_id", traceID,
				"user_id", c.Get("userID"),
			)

			// Log errors at error level
			if err != nil {
				logger.Error("Request Error",
					"error", err.Error(),
					"path", c.Request().RequestURI,
					"method", c.Request().Method,
					"trace_id", traceID,
				)
			}

			return err
		}
	}
}

// ErrorLoggerMiddleware logs panics and fatal errors
func ErrorLoggerMiddleware(logger *slog.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) (err error) {
			defer func() {
				if r := recover(); r != nil {
					logger.Error("Panic Recovered",
						"panic", r,
						"path", c.Request().RequestURI,
						"method", c.Request().Method,
						"trace_id", c.Get("traceID"),
					)
				}
			}()

			return next(c)
		}
	}
}
