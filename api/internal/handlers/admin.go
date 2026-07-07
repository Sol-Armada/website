package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/sol-armada/sol-bot/projects"
	"github.com/sol-armada/website/internal/dto"
	"github.com/sol-armada/website/internal/service"
)

type taskActivityResponse struct {
	ID      string `json:"id"`
	Summary string `json:"summary"`
	Time    string `json:"time,omitempty"`
}

type taskWithActivityResponse struct {
	*projects.Task
	Activity []taskActivityResponse `json:"activity"`
}

func (h *Handler) GetOverview(c echo.Context) error {
	result, err := service.GetOverviewStats(c.Request().Context())
	if err != nil {
		h.logger.Error("Failed to fetch admin overview", "error", err)
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "admin_overview_failed",
			Message: "Failed to fetch overview statistics",
		})
	}

	return c.JSON(http.StatusOK, result)
}
