package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"

	"github.com/sol-armada/website/internal/dto"
	"github.com/sol-armada/website/internal/service"
)

type MemberHandler struct {
	memberService *service.MemberService
	logger        *logrus.Logger
}

func NewMemberHandler(memberService *service.MemberService, logger *logrus.Logger) *MemberHandler {
	return &MemberHandler{
		memberService: memberService,
		logger:        logger,
	}
}

func (h *MemberHandler) GetDashboard(c echo.Context) error {
	memberID, _ := c.Get("user_id").(string)
	if memberID == "" {
		return c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error:   "unauthorized",
			Message: "Authentication required",
		})
	}

	result, err := h.memberService.GetDashboard(c.Request().Context(), memberID)
	if err != nil {
		h.logger.WithError(err).WithField("member_id", memberID).Error("Failed to fetch member dashboard")
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "member_dashboard_failed",
			Message: "Failed to fetch member dashboard",
		})
	}

	return c.JSON(http.StatusOK, result)
}

func (h *MemberHandler) GetProfile(c echo.Context) error {
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

	result, err := h.memberService.GetProfile(c.Request().Context(), memberID, username, email, roles)
	if err != nil {
		h.logger.WithError(err).WithField("member_id", memberID).Error("Failed to fetch member profile")
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "member_profile_failed",
			Message: "Failed to fetch member profile",
		})
	}

	return c.JSON(http.StatusOK, result)
}
