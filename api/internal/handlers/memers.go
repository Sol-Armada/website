package handlers

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/sol-armada/website/internal/dto"
	"github.com/sol-armada/website/internal/service"
)

func (h *Handler) GetMembers(c echo.Context) error {
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

	result, err := service.GetMembers(c.Request().Context(), limit, page, search)
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

func (h *Handler) GetMembersByAttendance(c echo.Context) error {
	attendanceId := c.Param("id")
	if attendanceId == "" {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid_request",
			Message: "Attendance Id is required",
		})
	}

	members, err := service.GetMembersByAttendance(c.Request().Context(), attendanceId)
	if err != nil {
		h.logger.Error("Failed to fetch members by attendance", "error", err)
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "members_fetch_failed",
			Message: "Failed to fetch members for the given attendance",
		})
	}

	return c.JSON(http.StatusOK, members)
}
