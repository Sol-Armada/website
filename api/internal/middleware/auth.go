package middleware

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/sol-armada/website/internal/auth"
)

// AuthMiddleware validates JWT tokens from cookies
type AuthMiddleware struct {
	tokenService  *auth.TokenService
	cookieService *auth.CookieService
}

// NewAuthMiddleware creates a new auth middleware
func NewAuthMiddleware(tokenService *auth.TokenService, cookieService *auth.CookieService) *AuthMiddleware {
	return &AuthMiddleware{
		tokenService:  tokenService,
		cookieService: cookieService,
	}
}

// RequireAuth middleware ensures the request has a valid session
func (m *AuthMiddleware) RequireAuth(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Get session cookie
		token, err := m.cookieService.GetSessionCookie(c)
		if err != nil {
			return echo.NewHTTPError(http.StatusUnauthorized, "Authentication required")
		}

		// Validate token
		claims, err := m.tokenService.ValidateToken(token)
		if err != nil {
			m.cookieService.ClearSessionCookie(c)
			return echo.NewHTTPError(http.StatusUnauthorized, "Invalid or expired session")
		}

		// Store claims in context for handlers to use
		c.Set("user_id", claims.UserID)
		c.Set("discord_id", claims.DiscordID)
		c.Set("username", claims.Username)
		c.Set("email", claims.Email)
		c.Set("roles", claims.Roles)

		return next(c)
	}
}

// RequireRole middleware ensures the user has one of the required roles
func (m *AuthMiddleware) RequireRole(allowedRoles ...string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Get roles from context (set by RequireAuth)
			rolesInterface := c.Get("roles")
			if rolesInterface == nil {
				return echo.NewHTTPError(http.StatusForbidden, "No roles found")
			}

			userRoles, ok := rolesInterface.([]string)
			if !ok {
				return echo.NewHTTPError(http.StatusForbidden, "Invalid roles format")
			}

			// Check if user has any of the allowed roles
			for _, userRole := range userRoles {
				for _, allowedRole := range allowedRoles {
					if strings.EqualFold(userRole, allowedRole) {
						return next(c)
					}
				}
			}

			return echo.NewHTTPError(http.StatusForbidden, "Insufficient permissions")
		}
	}
}

// OptionalAuth middleware validates token if present but doesn't require it
func (m *AuthMiddleware) OptionalAuth(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		token, err := m.cookieService.GetSessionCookie(c)
		if err == nil {
			claims, err := m.tokenService.ValidateToken(token)
			if err == nil {
				// Store claims in context
				c.Set("user_id", claims.UserID)
				c.Set("discord_id", claims.DiscordID)
				c.Set("username", claims.Username)
				c.Set("email", claims.Email)
				c.Set("roles", claims.Roles)
			}
		}
		return next(c)
	}
}

// CSRFMiddleware validates CSRF tokens for state-changing operations
func (m *AuthMiddleware) CSRFMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Only check CSRF for state-changing methods
		method := c.Request().Method
		if method == http.MethodGet || method == http.MethodHead || method == http.MethodOptions {
			return next(c)
		}

		// Get CSRF token from cookie
		cookieToken, err := m.cookieService.GetCSRFCookie(c)
		if err != nil {
			return echo.NewHTTPError(http.StatusForbidden, "CSRF token missing")
		}

		// Get CSRF token from header
		headerToken := c.Request().Header.Get("X-CSRF-Token")
		if headerToken == "" {
			return echo.NewHTTPError(http.StatusForbidden, "CSRF token required")
		}

		// Compare tokens
		if cookieToken != headerToken {
			return echo.NewHTTPError(http.StatusForbidden, "CSRF token mismatch")
		}

		return next(c)
	}
}
