package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
)

func TestRateLimiterAllowsThenLimits(t *testing.T) {
	e := echo.New()
	rl := NewRateLimiter(1, 1)

	handler := rl.Middleware()(func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	})

	req1 := httptest.NewRequest(http.MethodGet, "/", nil)
	req1.Header.Set("X-Forwarded-For", "10.0.0.1")
	rec1 := httptest.NewRecorder()
	ctx1 := e.NewContext(req1, rec1)

	if err := handler(ctx1); err != nil {
		t.Fatalf("first request returned error: %v", err)
	}
	if rec1.Code != http.StatusOK {
		t.Fatalf("first request status = %d, expected %d", rec1.Code, http.StatusOK)
	}

	req2 := httptest.NewRequest(http.MethodGet, "/", nil)
	req2.Header.Set("X-Forwarded-For", "10.0.0.1")
	rec2 := httptest.NewRecorder()
	ctx2 := e.NewContext(req2, rec2)

	err := handler(ctx2)
	if err == nil {
		t.Fatalf("second request expected rate-limit error, got nil")
	}

	httpErr, ok := err.(*echo.HTTPError)
	if !ok {
		t.Fatalf("expected *echo.HTTPError, got %T", err)
	}
	if httpErr.Code != http.StatusTooManyRequests {
		t.Fatalf("status = %d, expected %d", httpErr.Code, http.StatusTooManyRequests)
	}
}

func TestGetLimiterReturnsSameLimiterPerIP(t *testing.T) {
	rl := NewRateLimiter(5, 5)

	first := rl.getLimiter("127.0.0.1")
	second := rl.getLimiter("127.0.0.1")
	third := rl.getLimiter("127.0.0.2")

	if first != second {
		t.Fatalf("expected same limiter instance for same IP")
	}
	if first == third {
		t.Fatalf("expected different limiter instance for different IPs")
	}
}
