package handlers

import (
	"cmp"
	"context"
	"errors"
	"net/http"
	"slices"
	"strconv"
	"strings"

	"log/slog"

	"github.com/labstack/echo/v4"

	"github.com/sol-armada/website/internal/dto"
	"github.com/sol-armada/website/internal/service"
)

// AdminServiceInterface defines the interface for admin operations
type AdminServiceInterface interface {
	GetOverviewStats(context.Context) (*service.AdminOverviewStats, error)
	GetAttendanceRecords(context.Context, int, int, string) ([]service.AttendanceRecord, error)
	GetTokenLedger(context.Context, int, int, string) ([]service.TokenTransaction, error)
	GetTokenLedgerAnalytics(context.Context) (*service.TokenLedgerAnalytics, error)
	GetAttendanceAnalytics(context.Context) (*service.AttendanceAnalytics, error)
	GetMembers(context.Context, int, int, string) ([]service.MemberSummary, error)
	GetMembersByIds(context.Context, []string) (map[string]service.MemberSummary, error)
	CreateAttendanceRecord(context.Context, service.CreateAttendanceRecordInput) error
	GetAttendanceRecord(context.Context, string) (*service.AttendanceRecord, error)
	GetMembersByAttendance(context.Context, string) ([]service.MemberSummary, error)
	GetAttendanceEditPayload(context.Context, string) (*service.AttendanceEditPayload, error)
	UpdateAttendanceRecord(context.Context, string, service.UpdateAttendanceRecordInput) error
}

var _ AdminServiceInterface = (*service.AdminService)(nil)

type AdminHandler struct {
	adminService  AdminServiceInterface
	configService *service.ConfigService
	logger        *slog.Logger
}

func NewAdminHandler(adminService AdminServiceInterface, configService *service.ConfigService, logger *slog.Logger) *AdminHandler {
	return &AdminHandler{
		adminService:  adminService,
		configService: configService,
		logger:        logger,
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

	search := c.QueryParam("search")

	result, err := h.adminService.GetAttendanceRecords(c.Request().Context(), limit, page, search)
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

	search := c.QueryParam("search")

	result, err := h.adminService.GetTokenLedger(c.Request().Context(), limit, page, search)
	if err != nil {
		h.logger.Error("Failed to fetch token ledger", "error", err)
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "token_ledger_failed",
			Message: "Failed to fetch token ledger",
		})
	}

	// collect ids to turn into names
	userIDs := make(map[string]struct{})
	for _, tx := range result {
		userIDs[tx.MemberID] = struct{}{}
	}

	memberIDs := make([]string, 0, len(userIDs))
	for id := range userIDs {
		memberIDs = append(memberIDs, id)
	}

	members, err := h.adminService.GetMembersByIds(c.Request().Context(), memberIDs)
	if err != nil {
		h.logger.Error("Failed to fetch member details", "error", err)
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "member_fetch_failed",
			Message: "Failed to fetch member details",
		})
	}

	memberNames := make(map[string]string)
	for id, member := range members {
		memberNames[id] = member.Username
	}

	// enrich transactions with member names
	for i, tx := range result {
		if name, ok := memberNames[tx.MemberID]; ok {
			result[i].MemberName = tx.MemberID
			if name != "" {
				result[i].MemberName = name
			}
		}
	}

	slices.SortFunc(result, func(a, b service.TokenTransaction) int {
		if a.CreatedAt.After(b.CreatedAt) {
			return -1
		} else if a.CreatedAt.Before(b.CreatedAt) {
			return 1
		}
		return cmp.Compare(a.MemberName, b.MemberName)
	})

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

func (h *AdminHandler) GetAttendanceAnalytics(c echo.Context) error {
	roles, _ := c.Get("roles").([]string)
	if !hasRole(roles, "admin") {
		return c.JSON(http.StatusForbidden, dto.ErrorResponse{
			Error:   "forbidden",
			Message: "Admin access required",
		})
	}

	result, err := h.adminService.GetAttendanceAnalytics(c.Request().Context())
	if err != nil {
		h.logger.Error("Failed to fetch attendance analytics", "error", err)
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "attendance_analytics_failed",
			Message: "Failed to fetch attendance analytics",
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

func (h *AdminHandler) GetAvailableAttendanceNames(c echo.Context) error {
	roles, _ := c.Get("roles").([]string)
	if !hasRole(roles, "admin") {
		return c.JSON(http.StatusForbidden, dto.ErrorResponse{
			Error:   "forbidden",
			Message: "Admin access required",
		})
	}

	attendanceNames, err := h.configService.GetAvailableAttendanceNames()
	if err != nil {
		h.logger.Error("Failed to fetch attendance names", "error", err)
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "attendance_names_failed",
			Message: "Failed to fetch attendance names",
		})
	}

	return c.JSON(http.StatusOK, attendanceNames)
}

func (h *AdminHandler) CreateAttendanceRecord(c echo.Context) error {
	roles, _ := c.Get("roles").([]string)
	if !hasRole(roles, "admin") {
		return c.JSON(http.StatusForbidden, dto.ErrorResponse{
			Error:   "forbidden",
			Message: "Admin access required",
		})
	}

	// get the body
	req := service.CreateAttendanceRecordInput{}
	if err := c.Bind(&req); err != nil {
		h.logger.Error("Failed to bind request body", "error", err)
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
		})
	}

	if err := h.adminService.CreateAttendanceRecord(c.Request().Context(), req); err != nil {
		h.logger.Error("Failed to create attendance record", "error", err)
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "attendance_creation_failed",
			Message: "Failed to create attendance record",
		})
	}

	return c.JSON(http.StatusOK, map[string]any{
		"message": "Attendance record created successfully",
	})
}

func (h *AdminHandler) GetAttendanceRecord(c echo.Context) error {
	roles, _ := c.Get("roles").([]string)
	if !hasRole(roles, "admin") {
		return c.JSON(http.StatusForbidden, dto.ErrorResponse{
			Error:   "forbidden",
			Message: "Admin access required",
		})
	}

	id := c.Param("id")
	if id == "" {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid_request",
			Message: "Attendance record ID is required",
		})
	}

	record, err := h.adminService.GetAttendanceRecord(c.Request().Context(), id)
	if err != nil {
		h.logger.Error("Failed to fetch attendance record", "error", err)
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "attendance_fetch_failed",
			Message: "Failed to fetch attendance record",
		})
	}

	if record == nil {
		return c.JSON(http.StatusNotFound, dto.ErrorResponse{
			Error:   "not_found",
			Message: "Attendance record not found",
		})
	}

	return c.JSON(http.StatusOK, record)
}

func (h *AdminHandler) GetMembersByAttendance(c echo.Context) error {
	roles, _ := c.Get("roles").([]string)
	if !hasRole(roles, "admin") {
		return c.JSON(http.StatusForbidden, dto.ErrorResponse{
			Error:   "forbidden",
			Message: "Admin access required",
		})
	}

	attendanceId := c.Param("id")
	if attendanceId == "" {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid_request",
			Message: "Attendance ID is required",
		})
	}

	members, err := h.adminService.GetMembersByAttendance(c.Request().Context(), attendanceId)
	if err != nil {
		h.logger.Error("Failed to fetch members by attendance", "error", err)
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "members_fetch_failed",
			Message: "Failed to fetch members for the given attendance",
		})
	}

	return c.JSON(http.StatusOK, members)
}

func (h *AdminHandler) GetAttendanceEditPayload(c echo.Context) error {
	roles, _ := c.Get("roles").([]string)
	if !hasRole(roles, "admin") {
		return c.JSON(http.StatusForbidden, dto.ErrorResponse{
			Error:   "forbidden",
			Message: "Admin access required",
		})
	}

	id := strings.TrimSpace(c.Param("id"))
	if id == "" {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid_request",
			Message: "Attendance record ID is required",
		})
	}

	payload, err := h.adminService.GetAttendanceEditPayload(c.Request().Context(), id)
	if err != nil {
		h.logger.Error("Failed to fetch attendance edit payload", "attendanceId", id, "error", err)
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "attendance_fetch_failed",
			Message: "Failed to fetch attendance edit payload",
		})
	}

	if payload == nil {
		return c.JSON(http.StatusNotFound, dto.ErrorResponse{
			Error:   "not_found",
			Message: "Attendance record not found",
		})
	}

	return c.JSON(http.StatusOK, payload)
}

func (h *AdminHandler) UpdateAttendanceRecord(c echo.Context) error {
	roles, _ := c.Get("roles").([]string)
	if !hasRole(roles, "admin") {
		return c.JSON(http.StatusForbidden, dto.ErrorResponse{
			Error:   "forbidden",
			Message: "Admin access required",
		})
	}

	id := strings.TrimSpace(c.Param("id"))
	if id == "" {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid_request",
			Message: "Attendance record ID is required",
		})
	}

	req := service.UpdateAttendanceRecordInput{}
	if err := c.Bind(&req); err != nil {
		h.logger.Error("Failed to bind attendance update request", "attendanceId", id, "error", err)
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
		})
	}

	availableNames, err := h.configService.GetAvailableAttendanceNames()
	if err != nil {
		h.logger.Error("Failed to fetch available attendance names", "attendanceId", id, "error", err)
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "attendance_names_failed",
			Message: "Failed to validate attendance name",
		})
	}

	if req.Name != "" && !slices.Contains(availableNames, req.Name) {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid_request",
			Message: "Attendance name must be selected from the approved list",
		})
	}

	if err := h.adminService.UpdateAttendanceRecord(c.Request().Context(), id, req); err != nil {
		h.logger.Error("Failed to update attendance record", "attendanceId", id, "error", err)

		switch {
		case errors.Is(err, service.ErrInvalidAttendanceInput):
			return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
				Error:   "invalid_request",
				Message: err.Error(),
			})
		case errors.Is(err, service.ErrAttendanceRecordNotFound):
			return c.JSON(http.StatusNotFound, dto.ErrorResponse{
				Error:   "not_found",
				Message: "Attendance record not found",
			})
		default:
			return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
				Error:   "attendance_update_failed",
				Message: "Failed to update attendance record",
			})
		}
	}

	payload, err := h.adminService.GetAttendanceEditPayload(c.Request().Context(), id)
	if err != nil {
		h.logger.Error("Failed to reload attendance after update", "attendanceId", id, "error", err)
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "attendance_fetch_failed",
			Message: "Attendance updated but reloading failed",
		})
	}

	return c.JSON(http.StatusOK, payload)
}

func hasRole(roles []string, role string) bool {
	return slices.Contains(roles, role)
}
