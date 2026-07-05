package middleware

import (
	"net/http"
	"slices"

	"github.com/labstack/echo/v4"
	"github.com/sol-armada/website/internal/dto"
)

// RequireAdmin middleware ensures the request has admin role
func RequireAdmin(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		roles, _ := c.Get("roles").([]string)
		if !slices.Contains(roles, "admin") {
			return c.JSON(http.StatusForbidden, dto.ErrorResponse{
				Error:   "forbidden",
				Message: "Admin access required",
			})
		}
		return next(c)
	}
}
