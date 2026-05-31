package middleware

import (
	"time"

	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
)

// LoggingMiddleware logs HTTP requests with structured fields
func LoggingMiddleware(logger *logrus.Logger) echo.MiddlewareFunc {
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
			logger.WithFields(logrus.Fields{
				"method":      c.Request().Method,
				"path":        c.Request().RequestURI,
				"status":      c.Response().Status,
				"latency_ms":  latency,
				"user_agent":  c.Request().UserAgent(),
				"remote_ip":   c.RealIP(),
				"trace_id":    traceID,
				"user_id":     c.Get("userID"),
			}).Info("HTTP Request")

			// Log errors at error level
			if err != nil {
				logger.WithFields(logrus.Fields{
					"error":     err.Error(),
					"path":      c.Request().RequestURI,
					"method":    c.Request().Method,
					"trace_id":  traceID,
				}).Error("Request Error")
			}

			return err
		}
	}
}

// ErrorLoggerMiddleware logs panics and fatal errors
func ErrorLoggerMiddleware(logger *logrus.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) (err error) {
			defer func() {
				if r := recover(); r != nil {
					logger.WithFields(logrus.Fields{
						"panic":     r,
						"path":      c.Request().RequestURI,
						"method":    c.Request().Method,
						"trace_id":  c.Get("traceID"),
					}).Error("Panic Recovered")
				}
			}()

			return next(c)
		}
	}
}
