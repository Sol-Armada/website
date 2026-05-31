package service

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"log/slog"

	"github.com/sol-armada/sol-bot/attendance"
	"github.com/sol-armada/sol-bot/members"
	"github.com/sol-armada/sol-bot/tokens"
)

type MemberActivity struct {
	Type  string `json:"type"`
	Title string `json:"title"`
	Date  string `json:"date"`
}

type MemberDashboardData struct {
	Attendance     int              `json:"attendance"`
	Tokens         int              `json:"tokens"`
	Rank           string           `json:"rank"`
	RecentActivity []MemberActivity `json:"recentActivity"`
}

type MemberProfileData struct {
	ID              string   `json:"id"`
	DiscordID       string   `json:"discordID"`
	Username        string   `json:"username"`
	Email           string   `json:"email"`
	Roles           []string `json:"roles"`
	Rank            string   `json:"rank"`
	AttendanceCount int      `json:"attendanceCount"`
	TokensBalance   int      `json:"tokensBalance"`
	MemberSince     string   `json:"memberSince,omitempty"`
	RSIHandle       string   `json:"rsiHandle,omitempty"`
}

type MemberService struct {
	logger *slog.Logger
}

func NewMemberService(logger *slog.Logger) *MemberService {
	return &MemberService{logger: logger}
}

func (s *MemberService) GetDashboard(_ context.Context, memberID string) (*MemberDashboardData, error) {
	attendanceCount, err := attendance.GetMemberAttendanceCount(memberID)
	if err != nil {
		return nil, fmt.Errorf("failed to load attendance count: %w", err)
	}

	balance, err := tokens.GetBalanceByMemberId(memberID)
	if err != nil {
		return nil, fmt.Errorf("failed to load token balance: %w", err)
	}

	rankName := "Recruit"
	member, err := members.Get(memberID)
	if err != nil {
		if !errors.Is(err, members.MemberNotFound) {
			return nil, fmt.Errorf("failed to load member profile: %w", err)
		}
	} else if member != nil {
		if r := member.Rank.String(); r != "" {
			rankName = r
		}
	}

	recentActivity, err := s.getRecentTokenActivity(memberID)
	if err != nil {
		s.logger.Warn("Failed to load recent token activity", "error", err, "member_id", memberID)
		recentActivity = []MemberActivity{}
	}

	return &MemberDashboardData{
		Attendance:     attendanceCount,
		Tokens:         balance,
		Rank:           rankName,
		RecentActivity: recentActivity,
	}, nil
}

func (s *MemberService) GetProfile(_ context.Context, memberID, fallbackUsername, fallbackEmail string, roles []string) (*MemberProfileData, error) {
	attendanceCount, err := attendance.GetMemberAttendanceCount(memberID)
	if err != nil {
		return nil, fmt.Errorf("failed to load attendance count: %w", err)
	}

	balance, err := tokens.GetBalanceByMemberId(memberID)
	if err != nil {
		return nil, fmt.Errorf("failed to load token balance: %w", err)
	}

	result := &MemberProfileData{
		ID:              memberID,
		DiscordID:       memberID,
		Username:        fallbackUsername,
		Email:           fallbackEmail,
		Roles:           roles,
		Rank:            "Recruit",
		AttendanceCount: attendanceCount,
		TokensBalance:   balance,
	}

	member, err := members.Get(memberID)
	if err != nil {
		if !errors.Is(err, members.MemberNotFound) {
			return nil, fmt.Errorf("failed to load member profile: %w", err)
		}
		return result, nil
	}

	if member != nil {
		if member.Name != "" {
			result.Username = member.Name
		}
		if r := member.Rank.String(); r != "" {
			result.Rank = r
		}
		if !member.MemberSince.IsZero() {
			result.MemberSince = member.MemberSince.Format(time.RFC3339)
		}
		if member.RsiInfo != nil {
			result.RSIHandle = member.RsiInfo.Handle
		}
	}

	return result, nil
}

func (s *MemberService) getRecentTokenActivity(memberID string) ([]MemberActivity, error) {
	records, err := tokens.GetAll()
	if err != nil {
		return nil, err
	}

	filtered := make([]tokens.TokenRecord, 0)
	for _, record := range records {
		if record.MemberId == memberID {
			filtered = append(filtered, record)
		}
	}

	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].CreatedAt.After(filtered[j].CreatedAt)
	})

	limit := min(len(filtered), 5)

	result := make([]MemberActivity, 0, limit)
	for i := 0; i < limit; i++ {
		record := filtered[i]
		title := fmt.Sprintf("%+d tokens - %s", record.Amount, strings.TrimSpace(string(record.Reason)))
		result = append(result, MemberActivity{
			Type:  "token",
			Title: title,
			Date:  record.CreatedAt.Format(time.RFC3339),
		})
	}

	return result, nil
}
