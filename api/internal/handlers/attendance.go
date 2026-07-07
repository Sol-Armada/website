package handlers

import (
	"errors"
	"net/http"
	"slices"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/sol-armada/website/internal/dto"
	"github.com/sol-armada/website/internal/service"
)

func (h *Handler) GetAttendanceNames(c echo.Context) error {
	attendanceNames, err := service.GetAvailableAttendanceNames()
	if err != nil {
		h.logger.Error("Failed to fetch attendance names", "error", err)
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "attendance_names_failed",
			Message: "Failed to fetch attendance names",
		})
	}

	return c.JSON(http.StatusOK, attendanceNames)
}

func (h *Handler) CreateAttendanceName(c echo.Context) error {
	var req struct {
		Name string `json:"name"`
	}

	if err := c.Bind(&req); err != nil {
		h.logger.Error("Failed to bind create attendance name request", "error", err)
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
		})
	}

	name := strings.TrimSpace(req.Name)
	if name == "" {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid_request",
			Message: "Attendance name is required",
		})
	}

	if err := service.CreateAttendanceName(name); err != nil {
		h.logger.Error("Failed to create attendance name", "name", name, "error", err)
		if strings.Contains(strings.ToLower(err.Error()), "exist") {
			return c.JSON(http.StatusConflict, dto.ErrorResponse{
				Error:   "attendance_name_exists",
				Message: "Attendance name already exists",
			})
		}

		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "attendance_name_create_failed",
			Message: "Failed to create attendance name",
		})
	}

	return c.JSON(http.StatusOK, map[string]any{
		"message": "Attendance name created successfully",
		"name":    name,
	})
}

func (h *Handler) GetAttendance(c echo.Context) error {
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

	result, err := service.GetAttendanceRecords(c.Request().Context(), limit, page, search)
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

func (h *Handler) GetAttendanceAnalytics(c echo.Context) error {
	result, err := service.GetAttendanceAnalytics(c.Request().Context())
	if err != nil {
		h.logger.Error("Failed to fetch attendance analytics", "error", err)
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "attendance_analytics_failed",
			Message: "Failed to fetch attendance analytics",
		})
	}

	return c.JSON(http.StatusOK, result)
}

func (h *Handler) DeleteAttendanceName(c echo.Context) error {
	var req struct {
		Name string `json:"name"`
	}

	if err := c.Bind(&req); err != nil {
		h.logger.Error("Failed to bind delete attendance name request", "error", err)
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
		})
	}

	name := strings.TrimSpace(req.Name)
	if name == "" {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid_request",
			Message: "Attendance name is required",
		})
	}

	if err := service.DeleteAttendanceName(name); err != nil {
		h.logger.Error("Failed to delete attendance name", "name", name, "error", err)
		if strings.Contains(strings.ToLower(err.Error()), "not found") {
			return c.JSON(http.StatusNotFound, dto.ErrorResponse{
				Error:   "attendance_name_not_found",
				Message: "Attendance name not found",
			})
		}

		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "attendance_name_delete_failed",
			Message: "Failed to delete attendance name",
		})
	}

	return c.JSON(http.StatusOK, map[string]any{
		"message": "Attendance name deleted successfully",
		"name":    name,
	})
}

func (h *Handler) CreateAttendanceRecord(c echo.Context) error {
	// get the body
	req := service.CreateAttendanceRecordInput{}
	if err := c.Bind(&req); err != nil {
		h.logger.Error("Failed to bind request body", "error", err)
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
		})
	}

	if err := service.CreateAttendanceRecord(c.Request().Context(), req); err != nil {
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

func (h *Handler) GetAttendanceRecord(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid_request",
			Message: "Attendance record Id is required",
		})
	}

	record, err := service.GetAttendanceRecord(c.Request().Context(), id)
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

func (h *Handler) UpdateAttendanceRecord(c echo.Context) error {
	id := strings.TrimSpace(c.Param("id"))
	if id == "" {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid_request",
			Message: "Attendance record Id is required",
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

	availableNames, err := service.GetAvailableAttendanceNames()
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

	if err := service.UpdateAttendanceRecord(c.Request().Context(), id, req); err != nil {
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

	payload, err := service.GetAttendanceEditPayload(c.Request().Context(), id)
	if err != nil {
		h.logger.Error("Failed to reload attendance after update", "attendanceId", id, "error", err)
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "attendance_fetch_failed",
			Message: "Attendance updated but reloading failed",
		})
	}

	return c.JSON(http.StatusOK, payload)
}

func (h *Handler) GetAttendanceEditPayload(c echo.Context) error {
	id := strings.TrimSpace(c.Param("id"))
	if id == "" {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid_request",
			Message: "Attendance record Id is required",
		})
	}

	payload, err := service.GetAttendanceEditPayload(c.Request().Context(), id)
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
