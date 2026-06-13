package service

import (
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

type PaginatedResponse struct {
	Records []TokenTransaction `json:"records"`
	Page    int                `json:"page"`
	Limit   int                `json:"limit"`
}

type MemberDashboardData struct {
	Attendance     int                `json:"attendance"`
	Tokens         int                `json:"tokens"`
	Rank           string             `json:"rank"`
	RecentActivity []MemberActivity   `json:"recentActivity"`
	TokenLedger    []TokenTransaction `json:"tokenLedger,omitempty"`
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
	Validated       bool     `json:"validated"`
}

type MemberService struct {
	logger *slog.Logger
}

func NewMemberService(logger *slog.Logger) *MemberService {
	return &MemberService{logger: logger}
}

func (s *MemberService) GetDashboard(memberID string) (*MemberDashboardData, error) {
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

	tokenRecords, err := tokens.ListByMemberId(memberID)
	if err != nil {
		s.logger.Warn("Failed to load token ledger", "error", err, "member_id", memberID)
		tokenRecords = []tokens.TokenRecord{}
	}

	// Transform token records into TokenTransaction format
	var attendanceMap map[string]*attendance.Attendance
	attendanceIds := make(map[string]bool)
	for _, record := range tokenRecords {
		if record.AttendanceId != nil {
			attendanceIds[*record.AttendanceId] = true
		}
	}

	if len(attendanceIds) > 0 {
		idSlice := make([]string, 0, len(attendanceIds))
		for id := range attendanceIds {
			idSlice = append(idSlice, id)
		}
		atts, err := attendance.ListByIds(idSlice)
		if err != nil {
			s.logger.Warn("Failed to load attendance details", "error", err, "member_id", memberID)
			atts = []*attendance.Attendance{}
		}
		attendanceMap = make(map[string]*attendance.Attendance)
		for _, att := range atts {
			attendanceMap[att.Id] = att
		}
	} else {
		attendanceMap = make(map[string]*attendance.Attendance)
	}

	tokenLedger := make([]TokenTransaction, 0, len(tokenRecords))
	for _, record := range tokenRecords {
		comment := ""
		if record.Comment != nil {
			comment = *record.Comment
		}
		giverId := ""
		if record.GiverId != nil {
			giverId = *record.GiverId
		}
		attendanceId := ""
		if record.AttendanceId != nil {
			attendanceId = *record.AttendanceId
		}

		attendanceName := ""
		if att, ok := attendanceMap[attendanceId]; ok && att != nil {
			attendanceName = att.Name
		}

		tokenLedger = append(tokenLedger, TokenTransaction{
			ID:             record.Id,
			MemberID:       record.MemberId,
			Amount:         record.Amount,
			Reason:         string(record.Reason),
			CreatedAt:      record.CreatedAt,
			Comment:        comment,
			GiverID:        giverId,
			AttendanceID:   attendanceId,
			AttendanceName: attendanceName,
		})
	}

	return &MemberDashboardData{
		Attendance:     attendanceCount,
		Tokens:         balance,
		Rank:           rankName,
		RecentActivity: recentActivity,
		TokenLedger:    tokenLedger,
	}, nil
}

func (s *MemberService) GetProfile(memberID, fallbackUsername, fallbackEmail string, roles []string) (*MemberProfileData, error) {
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
		Validated:       false,
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
		result.Validated = (member.ValidatedAt != nil && !member.ValidatedAt.IsZero())
	}

	return result, nil
}

func (s *MemberService) GetTokenLedger(memberID string, limit, page int) (*PaginatedResponse, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	if page < 1 {
		page = 1
	}

	tokenRecords, err := tokens.ListByMemberId(memberID)
	if err != nil {
		return nil, fmt.Errorf("failed to load token ledger: %w", err)
	}

	// Fetch attendance details
	attendanceIds := make(map[string]bool)
	for _, record := range tokenRecords {
		if record.AttendanceId != nil {
			attendanceIds[*record.AttendanceId] = true
		}
	}

	attendanceMap := make(map[string]*attendance.Attendance)
	if len(attendanceIds) > 0 {
		idSlice := make([]string, 0, len(attendanceIds))
		for id := range attendanceIds {
			idSlice = append(idSlice, id)
		}
		atts, err := attendance.ListByIds(idSlice)
		if err != nil {
			s.logger.Warn("Failed to load attendance details", "error", err, "member_id", memberID)
		}
		for _, att := range atts {
			attendanceMap[att.Id] = att
		}
	}

	// Transform token records into TokenTransaction format
	tokenLedger := make([]TokenTransaction, 0, len(tokenRecords))
	for _, record := range tokenRecords {
		comment := ""
		if record.Comment != nil {
			comment = *record.Comment
		}
		giverId := ""
		if record.GiverId != nil {
			giverId = *record.GiverId
		}
		attendanceId := ""
		if record.AttendanceId != nil {
			attendanceId = *record.AttendanceId
		}

		attendanceName := ""
		if att, ok := attendanceMap[attendanceId]; ok && att != nil {
			attendanceName = att.Name
		}

		tokenLedger = append(tokenLedger, TokenTransaction{
			ID:             record.Id,
			MemberID:       record.MemberId,
			Amount:         record.Amount,
			Reason:         string(record.Reason),
			CreatedAt:      record.CreatedAt,
			Comment:        comment,
			GiverID:        giverId,
			AttendanceID:   attendanceId,
			AttendanceName: attendanceName,
		})
	}

	// Paginate the results
	total := len(tokenLedger)
	start := (page - 1) * limit
	end := start + limit

	if start >= total {
		return &PaginatedResponse{
			Records: []TokenTransaction{},
			Page:    page,
			Limit:   limit,
		}, nil
	}

	if end > total {
		end = total
	}

	return &PaginatedResponse{
		Records: tokenLedger[start:end],
		Page:    page,
		Limit:   limit,
	}, nil
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
	for i := range limit {
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
