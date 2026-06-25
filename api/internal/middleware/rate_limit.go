package middleware

import (
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"golang.org/x/time/rate"
)

// RateLimiter manages per-IP rate limiting
type RateLimiter struct {
	limiters map[string]*rate.Limiter
	limit    rate.Limit
	burst    int
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(requestsPerSecond int, burst int) *RateLimiter {
	return &RateLimiter{
		limiters: make(map[string]*rate.Limiter),
		limit:    rate.Limit(requestsPerSecond),
		burst:    burst,
	}
}

// Middleware returns the rate limiting middleware
func (rl *RateLimiter) Middleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ip := c.RealIP()
			limiter := rl.getLimiter(ip)

			if !limiter.Allow() {
				return echo.NewHTTPError(
					http.StatusTooManyRequests,
					fmt.Sprintf("Rate limit exceeded for IP %s", ip),
				)
			}

			// Add rate limit headers
			c.Response().Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", rl.burst))
			c.Response().Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%d", int(limiter.Tokens())))
			c.Response().Header().Set("X-RateLimit-Reset", fmt.Sprintf("%d", time.Now().Add(time.Second).Unix()))

			return next(c)
		}
	}
}

// getLimiter gets or creates a limiter for an IP
func (rl *RateLimiter) getLimiter(ip string) *rate.Limiter {
	limiter, exists := rl.limiters[ip]
	if !exists {
		limiter = rate.NewLimiter(rl.limit, rl.burst)
		rl.limiters[ip] = limiter

		// Cleanup old limiters periodically
		go func() {
			time.Sleep(1 * time.Hour)
			delete(rl.limiters, ip)
		}()
	}

	return limiter
}

// AdminRateLimiter for stricter admin rate limiting
func AdminRateLimiter(requestsPerSecond int, burst int) echo.MiddlewareFunc {
	rl := NewRateLimiter(requestsPerSecond, burst)
	return rl.Middleware()
}
