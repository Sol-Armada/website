package handlers

import (
	"cmp"
	"net/http"
	"slices"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/sol-armada/website/internal/dto"
	"github.com/sol-armada/website/internal/service"
)

func (h *Handler) GetTokenLedger(c echo.Context) error {
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

	result, err := service.GetTokenLedger(c.Request().Context(), limit, page, search)
	if err != nil {
		h.logger.Error("Failed to fetch token ledger", "error", err)
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "token_ledger_failed",
			Message: "Failed to fetch token ledger",
		})
	}

	// collect ids to turn into names
	userIds := make(map[string]struct{})
	for _, tx := range result {
		userIds[tx.MemberID] = struct{}{}
	}

	memberIds := make([]string, 0, len(userIds))
	for id := range userIds {
		memberIds = append(memberIds, id)
	}

	members, err := service.GetMembersByIds(c.Request().Context(), memberIds)
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

func (h *Handler) GetTokenLedgerAnalytics(c echo.Context) error {
	result, err := service.GetTokenLedgerAnalytics(c.Request().Context())
	if err != nil {
		h.logger.Error("Failed to fetch token ledger analytics", "error", err)
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "token_analytics_failed",
			Message: "Failed to fetch token ledger analytics",
		})
	}

	return c.JSON(http.StatusOK, result)
}
