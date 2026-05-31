package auth

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
)

const (
	SessionCookieName = "sa_session"
	CSRFCookieName    = "sa_csrf"
)

// CookieService handles secure cookie operations
type CookieService struct {
	isProduction bool
	domain       string
}

// NewCookieService creates a new cookie service
func NewCookieService(environment string, domain string) *CookieService {
	return &CookieService{
		isProduction: environment == "production",
		domain:       domain,
	}
}

// SetSessionCookie sets a secure session cookie with JWT token
func (s *CookieService) SetSessionCookie(c echo.Context, token string, maxAge int) {
	cookie := &http.Cookie{
		Name:     SessionCookieName,
		Value:    token,
		Path:     "/",
		MaxAge:   maxAge,
		HttpOnly: true,
		Secure:   s.isProduction,
		SameSite: http.SameSiteLaxMode,
	}
	
	if s.domain != "" {
		cookie.Domain = s.domain
	}
	
	c.SetCookie(cookie)
}

// SetCSRFCookie sets a CSRF token cookie
func (s *CookieService) SetCSRFCookie(c echo.Context, token string, maxAge int) {
	cookie := &http.Cookie{
		Name:     CSRFCookieName,
		Value:    token,
		Path:     "/",
		MaxAge:   maxAge,
		HttpOnly: false, // JavaScript needs to read this
		Secure:   s.isProduction,
		SameSite: http.SameSiteStrictMode,
	}
	
	if s.domain != "" {
		cookie.Domain = s.domain
	}
	
	c.SetCookie(cookie)
}

// ClearSessionCookie removes the session cookie
func (s *CookieService) ClearSessionCookie(c echo.Context) {
	cookie := &http.Cookie{
		Name:     SessionCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   s.isProduction,
		SameSite: http.SameSiteLaxMode,
	}
	
	c.SetCookie(cookie)
}

// ClearCSRFCookie removes the CSRF cookie
func (s *CookieService) ClearCSRFCookie(c echo.Context) {
	cookie := &http.Cookie{
		Name:     CSRFCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: false,
		Secure:   s.isProduction,
		SameSite: http.SameSiteStrictMode,
	}
	
	c.SetCookie(cookie)
}

// GetSessionCookie retrieves the session cookie value
func (s *CookieService) GetSessionCookie(c echo.Context) (string, error) {
	cookie, err := c.Cookie(SessionCookieName)
	if err != nil {
		return "", err
	}
	return cookie.Value, nil
}

// GetCSRFCookie retrieves the CSRF cookie value
func (s *CookieService) GetCSRFCookie(c echo.Context) (string, error) {
	cookie, err := c.Cookie(CSRFCookieName)
	if err != nil {
		return "", err
	}
	return cookie.Value, nil
}

// GenerateCSRFToken generates a random CSRF token
func GenerateCSRFToken() string {
	// Use crypto/rand for secure random token generation
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, 32)
	for i := range b {
		b[i] = charset[time.Now().UnixNano()%int64(len(charset))]
	}
	return string(b)
}
