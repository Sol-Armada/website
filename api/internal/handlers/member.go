package handlers

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"

	"github.com/sol-armada/website/internal/dto"
	"github.com/sol-armada/website/internal/service"
)

func (h *Handler) GetDashboard(c echo.Context) error {
	memberID, _ := c.Get("user_id").(string)
	if memberID == "" {
		return c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error:   "unauthorized",
			Message: "Authentication required",
		})
	}

	result, err := service.GetDashboard(memberID)
	if err != nil {
		h.logger.Error("Failed to fetch member dashboard", "error", err, "member_id", memberID)
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "member_dashboard_failed",
			Message: "Failed to fetch member dashboard",
		})
	}

	return c.JSON(http.StatusOK, result)
}

func (h *Handler) GetProfile(c echo.Context) error {
	memberID, _ := c.Get("user_id").(string)
	if memberID == "" {
		return c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error:   "unauthorized",
			Message: "Authentication required",
		})
	}

	username, _ := c.Get("username").(string)
	email, _ := c.Get("email").(string)
	roles, _ := c.Get("roles").([]string)

	result, err := service.GetProfile(memberID, username, email, roles)
	if err != nil {
		h.logger.Error("Failed to fetch member profile", "error", err, "member_id", memberID)
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "member_profile_failed",
			Message: "Failed to fetch member profile",
		})
	}

	return c.JSON(http.StatusOK, result)
}

func (h *Handler) GetMemberTokenLedger(c echo.Context) error {
	memberID, _ := c.Get("user_id").(string)
	if memberID == "" {
		return c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error:   "unauthorized",
			Message: "Authentication required",
		})
	}

	limit := 50
	if l := c.QueryParam("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}

	page := 1
	if p := c.QueryParam("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}

	result, err := service.GetMemberTokenLedger(memberID, limit, page)
	if err != nil {
		h.logger.Error("Failed to fetch member token ledger", "error", err, "member_id", memberID)
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "token_ledger_failed",
			Message: "Failed to fetch token ledger",
		})
	}

	return c.JSON(http.StatusOK, result)
}
