package service

import (
	"context"
	"fmt"
	"sort"
	"time"

	"log/slog"

	"github.com/sol-armada/sol-bot/attendance"
	"github.com/sol-armada/sol-bot/members"
	"github.com/sol-armada/sol-bot/tokens"
)

type AdminOverviewStats struct {
	TotalMembers      int `json:"totalMembers"`
	TotalEvents       int `json:"totalEvents"`
	TotalTokens       int `json:"totalTokens"`
	ActiveThisMonth   int `json:"activeThisMonth"`
	UniqueAttendees   int `json:"uniqueAttendees"`
	AverageAttendance int `json:"averageAttendance"`
}

type AttendanceRecord struct {
	ID               string    `json:"id"`
	Name             string    `json:"name"`
	SubmittedBy      string    `json:"submittedBy"`
	ParticipantCount int       `json:"participantCount"`
	Recorded         bool      `json:"recorded"`
	Successful       bool      `json:"successful"`
	DateCreated      time.Time `json:"dateCreated"`
}

type TokenTransaction struct {
	ID           string    `json:"id"`
	MemberID     string    `json:"memberId"`
	Amount       int       `json:"amount"`
	Reason       string    `json:"reason"`
	CreatedAt    time.Time `json:"createdAt"`
	Comment      string    `json:"comment,omitempty"`
	GiverID      string    `json:"giverId,omitempty"`
	AttendanceID string    `json:"attendanceId,omitempty"`
}

type TokenPeriodAnalytics struct {
	WindowStart                   time.Time `json:"windowStart"`
	WindowEnd                     time.Time `json:"windowEnd"`
	TotalEarnings                 int       `json:"totalEarnings"`
	TotalSpending                 int       `json:"totalSpending"`
	NetAmount                     int       `json:"netAmount"`
	AverageEarningPerMember       float64   `json:"averageEarningPerMember"`
	AverageSpendingPerMember      float64   `json:"averageSpendingPerMember"`
	AverageEarningPerTransaction  float64   `json:"averageEarningPerTransaction"`
	AverageSpendingPerTransaction float64   `json:"averageSpendingPerTransaction"`
	EarningTransactionCount       int       `json:"earningTransactionCount"`
	SpendingTransactionCount      int       `json:"spendingTransactionCount"`
	EarningMemberCount            int       `json:"earningMemberCount"`
	SpendingMemberCount           int       `json:"spendingMemberCount"`
}

type TokenReasonAggregation struct {
	Reason           string `json:"reason"`
	TransactionCount int    `json:"transactionCount"`
	NetAmount        int    `json:"netAmount"`
	TotalEarnings    int    `json:"totalEarnings"`
	TotalSpending    int    `json:"totalSpending"`
}

type TokenLedgerAnalytics struct {
	Week    TokenPeriodAnalytics     `json:"week"`
	Month   TokenPeriodAnalytics     `json:"month"`
	Reasons []TokenReasonAggregation `json:"reasons"`
}

type MemberSummary struct {
	ID           string `json:"id"`
	Username     string `json:"username"`
	Rank         string `json:"rank"`
	Attendance   int    `json:"attendance"`
	TokenBalance int    `json:"tokenBalance"`
	RSIHandle    string `json:"rsiHandle,omitempty"`
}

type AdminService struct {
	logger *slog.Logger
}

func NewAdminService(logger *slog.Logger) *AdminService {
	return &AdminService{logger: logger}
}

func (s *AdminService) GetOverviewStats(_ context.Context) (*AdminOverviewStats, error) {
	// Get all attendance records to count events
	allAttendance, err := attendance.List(10000, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch attendance records: %w", err)
	}

	// Count unique attendees in past 30 days
	uniqueAttendees, err := attendance.GetUniqueMemberCount(30)
	if err != nil {
		return nil, fmt.Errorf("failed to count unique attendees: %w", err)
	}

	// Get all members
	memberIDs, err := members.GetStoredMemberIDs()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch member IDs: %w", err)
	}

	// Get total tokens distributed
	allTokens, err := tokens.GetAll()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch tokens: %w", err)
	}

	totalTokens := 0
	for _, token := range allTokens {
		totalTokens += token.Amount
	}

	// Calculate average attendance (total participants / total events)
	avgAttendance := 0
	if len(allAttendance) > 0 {
		totalParticipants := 0
		for _, att := range allAttendance {
			participants, err := att.Participants()
			if err == nil {
				totalParticipants += len(participants)
			}
		}
		avgAttendance = totalParticipants / len(allAttendance)
	}

	return &AdminOverviewStats{
		TotalMembers:      len(memberIDs),
		TotalEvents:       len(allAttendance),
		TotalTokens:       totalTokens,
		ActiveThisMonth:   uniqueAttendees,
		UniqueAttendees:   uniqueAttendees,
		AverageAttendance: avgAttendance,
	}, nil
}

func (s *AdminService) GetAttendanceRecords(_ context.Context, limit, page int) ([]AttendanceRecord, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	if page < 1 {
		page = 1
	}

	attendanceList, err := attendance.List(limit, page)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch attendance records: %w", err)
	}

	result := make([]AttendanceRecord, 0, len(attendanceList))
	for _, att := range attendanceList {
		submittedBy := ""
		if att.SubmittedBy != nil {
			submittedBy = att.SubmittedBy.Name
		}

		participants, err := att.Participants()
		participantCount := 0
		if err == nil {
			participantCount = len(participants)
		}

		result = append(result, AttendanceRecord{
			ID:               att.Id,
			Name:             att.Name,
			SubmittedBy:      submittedBy,
			ParticipantCount: participantCount,
			Recorded:         att.Recorded,
			Successful:       att.Successful,
			DateCreated:      att.DateCreated,
		})
	}

	return result, nil
}

func (s *AdminService) GetTokenLedger(_ context.Context, limit, page int) ([]TokenTransaction, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	if page < 1 {
		page = 1
	}

	allTokens, err := tokens.GetAll()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch token records: %w", err)
	}

	// Sort by newest first
	for i, j := 0, len(allTokens)-1; i < j; i, j = i+1, j-1 {
		allTokens[i], allTokens[j] = allTokens[j], allTokens[i]
	}

	// Paginate
	start := (page - 1) * limit
	end := start + limit
	if start > len(allTokens) {
		start = len(allTokens)
	}
	if end > len(allTokens) {
		end = len(allTokens)
	}

	result := make([]TokenTransaction, 0)
	for i := start; i < end; i++ {
		t := allTokens[i]
		comment := ""
		if t.Comment != nil {
			comment = *t.Comment
		}
		giverId := ""
		if t.GiverId != nil {
			giverId = *t.GiverId
		}
		attendanceId := ""
		if t.AttendanceId != nil {
			attendanceId = *t.AttendanceId
		}

		result = append(result, TokenTransaction{
			ID:           t.Id,
			MemberID:     t.MemberId,
			Amount:       t.Amount,
			Reason:       string(t.Reason),
			CreatedAt:    t.CreatedAt,
			Comment:      comment,
			GiverID:      giverId,
			AttendanceID: attendanceId,
		})
	}

	return result, nil
}

func (s *AdminService) GetTokenLedgerAnalytics(_ context.Context) (*TokenLedgerAnalytics, error) {
	allTokens, err := tokens.GetAll()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch token records for analytics: %w", err)
	}

	now := time.Now().UTC()
	weekStart := now.AddDate(0, 0, -7)
	monthStart := now.AddDate(0, 0, -30)

	return &TokenLedgerAnalytics{
		Week:    calculateTokenPeriodAnalytics(allTokens, weekStart, now),
		Month:   calculateTokenPeriodAnalytics(allTokens, monthStart, now),
		Reasons: calculateReasonAggregation(allTokens),
	}, nil
}

func calculateTokenPeriodAnalytics(allTokens []tokens.TokenRecord, start, end time.Time) TokenPeriodAnalytics {
	result := TokenPeriodAnalytics{
		WindowStart: start,
		WindowEnd:   end,
	}

	earningMembers := make(map[string]struct{})
	spendingMembers := make(map[string]struct{})

	for _, token := range allTokens {
		createdAt := token.CreatedAt.UTC()
		if createdAt.Before(start) || createdAt.After(end) {
			continue
		}

		result.NetAmount += token.Amount

		if token.Amount > 0 {
			result.TotalEarnings += token.Amount
			result.EarningTransactionCount += 1
			earningMembers[token.MemberId] = struct{}{}
			continue
		}

		if token.Amount < 0 {
			result.TotalSpending += -token.Amount
			result.SpendingTransactionCount += 1
			spendingMembers[token.MemberId] = struct{}{}
		}
	}

	result.EarningMemberCount = len(earningMembers)
	result.SpendingMemberCount = len(spendingMembers)

	if result.EarningMemberCount > 0 {
		result.AverageEarningPerMember = float64(result.TotalEarnings) / float64(result.EarningMemberCount)
	}

	if result.SpendingMemberCount > 0 {
		result.AverageSpendingPerMember = float64(result.TotalSpending) / float64(result.SpendingMemberCount)
	}

	if result.EarningTransactionCount > 0 {
		result.AverageEarningPerTransaction = float64(result.TotalEarnings) / float64(result.EarningTransactionCount)
	}

	if result.SpendingTransactionCount > 0 {
		result.AverageSpendingPerTransaction = float64(result.TotalSpending) / float64(result.SpendingTransactionCount)
	}

	return result
}

func calculateReasonAggregation(allTokens []tokens.TokenRecord) []TokenReasonAggregation {
	totals := make(map[string]*TokenReasonAggregation)

	for _, token := range allTokens {
		reason := string(token.Reason)
		entry, exists := totals[reason]
		if !exists {
			entry = &TokenReasonAggregation{Reason: reason}
			totals[reason] = entry
		}

		entry.TransactionCount += 1
		entry.NetAmount += token.Amount

		if token.Amount > 0 {
			entry.TotalEarnings += token.Amount
		} else if token.Amount < 0 {
			entry.TotalSpending += -token.Amount
		}
	}

	result := make([]TokenReasonAggregation, 0, len(totals))
	for _, entry := range totals {
		result = append(result, *entry)
	}

	sort.Slice(result, func(i, j int) bool {
		if result[i].TransactionCount == result[j].TransactionCount {
			if result[i].NetAmount == result[j].NetAmount {
				return result[i].Reason < result[j].Reason
			}
			return result[i].NetAmount > result[j].NetAmount
		}
		return result[i].TransactionCount > result[j].TransactionCount
	})

	return result
}

func (s *AdminService) GetMembers(_ context.Context, limit, page int, search string) ([]MemberSummary, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	if page < 1 {
		page = 1
	}

	memberList, err := members.List(page)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch members: %w", err)
	}

	// Get all token balances
	allTokens, err := tokens.GetAll()
	if err != nil {
		// If this fails, we'll just use 0 balance
		allTokens = make([]tokens.TokenRecord, 0)
	}

	// Build balance map
	balances := make(map[string]int)
	for _, t := range allTokens {
		balances[t.MemberId] += t.Amount
	}

	result := make([]MemberSummary, 0)
	for _, m := range memberList {
		if m.Id == "" {
			continue
		}

		// Filter by search if provided
		if search != "" && !matchesSearch(m.Name, search) {
			continue
		}

		attendance, err := attendance.GetMemberAttendanceCount(m.Id)
		if err != nil {
			attendance = 0
		}

		rsiHandle := ""
		if m.RsiInfo != nil {
			rsiHandle = m.RsiInfo.Handle
		}

		result = append(result, MemberSummary{
			ID:           m.Id,
			Username:     m.Name,
			Rank:         m.Rank.String(),
			Attendance:   attendance,
			TokenBalance: balances[m.Id],
			RSIHandle:    rsiHandle,
		})

		if len(result) >= limit {
			break
		}
	}

	return result, nil
}

func matchesSearch(name, search string) bool {
	if search == "" {
		return true
	}
	return len(name) >= len(search) && name[:len(search)] == search[:len(name)]
}
