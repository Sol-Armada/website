package handlers

import (
	"context"
	"net/http"
	"slices"
	"strconv"

	"log/slog"

	"github.com/labstack/echo/v4"

	"github.com/sol-armada/website/internal/dto"
	"github.com/sol-armada/website/internal/service"
)

// AdminServiceInterface defines the interface for admin operations
type AdminServiceInterface interface {
	GetOverviewStats(context.Context) (*service.AdminOverviewStats, error)
	GetAttendanceRecords(context.Context, int, int) ([]service.AttendanceRecord, error)
	GetTokenLedger(context.Context, int, int) ([]service.TokenTransaction, error)
	GetTokenLedgerAnalytics(context.Context) (*service.TokenLedgerAnalytics, error)
	GetMembers(context.Context, int, int, string) ([]service.MemberSummary, error)
}

type AdminHandler struct {
	adminService AdminServiceInterface
	logger       *slog.Logger
}

func NewAdminHandler(adminService AdminServiceInterface, logger *slog.Logger) *AdminHandler {
	return &AdminHandler{
		adminService: adminService,
		logger:       logger,
	}
}

func (h *AdminHandler) GetOverview(c echo.Context) error {
	// Check if user has admin role
	roles, _ := c.Get("roles").([]string)
	if !hasRole(roles, "admin") {
		return c.JSON(http.StatusForbidden, dto.ErrorResponse{
			Error:   "forbidden",
			Message: "Admin access required",
		})
	}

	result, err := h.adminService.GetOverviewStats(c.Request().Context())
	if err != nil {
		h.logger.Error("Failed to fetch admin overview", "error", err)
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "admin_overview_failed",
			Message: "Failed to fetch overview statistics",
		})
	}

	return c.JSON(http.StatusOK, result)
}

func (h *AdminHandler) GetAttendance(c echo.Context) error {
	roles, _ := c.Get("roles").([]string)
	if !hasRole(roles, "admin") {
		return c.JSON(http.StatusForbidden, dto.ErrorResponse{
			Error:   "forbidden",
			Message: "Admin access required",
		})
	}

	limit := 50
	if l := c.QueryParam("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	page := 1
	if p := c.QueryParam("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}

	result, err := h.adminService.GetAttendanceRecords(c.Request().Context(), limit, page)
	if err != nil {
		h.logger.Error("Failed to fetch attendance records", "error", err)
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "attendance_fetch_failed",
			Message: "Failed to fetch attendance records",
		})
	}

	return c.JSON(http.StatusOK, map[string]any{
		"records": result,
		"page":    page,
		"limit":   limit,
	})
}

func (h *AdminHandler) GetTokenLedger(c echo.Context) error {
	roles, _ := c.Get("roles").([]string)
	if !hasRole(roles, "admin") {
		return c.JSON(http.StatusForbidden, dto.ErrorResponse{
			Error:   "forbidden",
			Message: "Admin access required",
		})
	}

	limit := 50
	if l := c.QueryParam("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	page := 1
	if p := c.QueryParam("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}

	result, err := h.adminService.GetTokenLedger(c.Request().Context(), limit, page)
	if err != nil {
		h.logger.Error("Failed to fetch token ledger", "error", err)
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "token_ledger_failed",
			Message: "Failed to fetch token ledger",
		})
	}

	return c.JSON(http.StatusOK, map[string]any{
		"records": result,
		"page":    page,
		"limit":   limit,
	})
}

func (h *AdminHandler) GetTokenLedgerAnalytics(c echo.Context) error {
	roles, _ := c.Get("roles").([]string)
	if !hasRole(roles, "admin") {
		return c.JSON(http.StatusForbidden, dto.ErrorResponse{
			Error:   "forbidden",
			Message: "Admin access required",
		})
	}

	result, err := h.adminService.GetTokenLedgerAnalytics(c.Request().Context())
	if err != nil {
		h.logger.Error("Failed to fetch token ledger analytics", "error", err)
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "token_analytics_failed",
			Message: "Failed to fetch token ledger analytics",
		})
	}

	return c.JSON(http.StatusOK, result)
}

func (h *AdminHandler) GetMembers(c echo.Context) error {
	roles, _ := c.Get("roles").([]string)
	if !hasRole(roles, "admin") {
		return c.JSON(http.StatusForbidden, dto.ErrorResponse{
			Error:   "forbidden",
			Message: "Admin access required",
		})
	}

	limit := 50
	if l := c.QueryParam("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	page := 1
	if p := c.QueryParam("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}

	search := c.QueryParam("search")

	result, err := h.adminService.GetMembers(c.Request().Context(), limit, page, search)
	if err != nil {
		h.logger.Error("Failed to fetch members", "error", err)
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "members_fetch_failed",
			Message: "Failed to fetch members",
		})
	}

	return c.JSON(http.StatusOK, map[string]any{
		"members": result,
		"page":    page,
		"limit":   limit,
	})
}

func hasRole(roles []string, role string) bool {
	return slices.Contains(roles, role)
}
