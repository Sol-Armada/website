package service

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
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
	MemberName   string    `json:"memberName,omitempty"`
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

func (s *AdminService) GetAttendanceRecords(_ context.Context, limit, page int, search string) ([]AttendanceRecord, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	if page < 1 {
		page = 1
	}

	normalizedSearch := normalizeSearch(search)

	listLimit := limit
	listPage := page
	if normalizedSearch != "" {
		listLimit = 10000
		listPage = 0
	}

	attendanceList, err := attendance.List(listLimit, listPage)
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

		if normalizedSearch != "" {
			if !matchesAnyField(normalizedSearch,
				att.Name,
				submittedBy,
				strconv.Itoa(participantCount),
				att.DateCreated.Format("2006-01-02"),
				att.DateCreated.Format(time.RFC3339),
				strconv.FormatBool(att.Recorded),
				strconv.FormatBool(att.Successful),
			) {
				continue
			}
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

	if normalizedSearch != "" {
		return paginateAttendanceRecords(result, limit, page), nil
	}

	return result, nil
}

func (s *AdminService) GetTokenLedger(_ context.Context, limit, page int, search string) ([]TokenTransaction, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	if page < 1 {
		page = 1
	}

	normalizedSearch := normalizeSearch(search)

	allTokens, err := tokens.GetAll()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch token records: %w", err)
	}

	// Sort by newest first
	for i, j := 0, len(allTokens)-1; i < j; i, j = i+1, j-1 {
		allTokens[i], allTokens[j] = allTokens[j], allTokens[i]
	}

	result := make([]TokenTransaction, 0, len(allTokens))
	for _, t := range allTokens {
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

		if normalizedSearch != "" {
			if !matchesAnyField(normalizedSearch,
				t.MemberId,
				strconv.Itoa(t.Amount),
				string(t.Reason),
				comment,
				giverId,
				attendanceId,
				t.CreatedAt.Format("2006-01-02"),
				t.CreatedAt.Format(time.RFC3339),
			) {
				continue
			}
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

	return paginateTokenTransactions(result, limit, page), nil
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

func (s *AdminService) GetMembers(ctx context.Context, limit, page int, search string) ([]MemberSummary, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}

	var memberList []members.Member
	var err error
	switch {
	case page == 0 || search != "":
		memberList, err = members.ListAll(ctx)
	case page > 0:
		memberList, err = members.List(ctx, page, limit)
	default:
		return nil, fmt.Errorf("invalid page number: %d", page)
	}
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
	balances := buildTokenBalanceMap(allTokens)

	normalizedSearch := normalizeSearch(search)

	result := make([]MemberSummary, 0)
	for _, m := range memberList {
		if m.Id == "" {
			continue
		}

		summary := buildMemberSummary(m, balances)

		if normalizedSearch != "" {
			if !matchesAnyField(normalizedSearch,
				summary.Username,
				summary.ID,
				summary.Rank,
				summary.RSIHandle,
				strconv.Itoa(summary.Attendance),
				strconv.Itoa(summary.TokenBalance),
			) {
				continue
			}
		}

		result = append(result, summary)

		if len(result) >= limit {
			break
		}
	}

	return result, nil
}

func (s *AdminService) GetMembersByIds(ctx context.Context, ids []string) (map[string]MemberSummary, error) {
	members, err := members.GetList(ids)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch members by ids: %w", err)
	}

	result := make(map[string]MemberSummary)
	for _, m := range members {
		summary := buildMemberSummary(*m, nil) // Assuming balances are not needed here
		result[m.Id] = summary
	}

	return result, nil
}

func (s *AdminService) GetMemberSummaryByID(_ context.Context, memberID string) (*MemberSummary, error) {
	if strings.TrimSpace(memberID) == "" {
		return nil, nil
	}

	member, err := members.Get(memberID)
	if err != nil {
		if errors.Is(err, members.MemberNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to fetch member %s: %w", memberID, err)
	}

	if member == nil {
		return nil, nil
	}

	allTokens, err := tokens.GetAll()
	if err != nil {
		allTokens = make([]tokens.TokenRecord, 0)
	}

	balances := buildTokenBalanceMap(allTokens)
	summary := buildMemberSummary(*member, balances)
	return &summary, nil
}

func buildTokenBalanceMap(allTokens []tokens.TokenRecord) map[string]int {
	balances := make(map[string]int)
	for _, token := range allTokens {
		balances[token.MemberId] += token.Amount
	}
	return balances
}

func buildMemberSummary(member members.Member, balances map[string]int) MemberSummary {
	attendanceCount, err := attendance.GetMemberAttendanceCount(member.Id)
	if err != nil {
		attendanceCount = 0
	}

	rsiHandle := ""
	if member.RsiInfo != nil {
		rsiHandle = member.RsiInfo.Handle
	}

	summary := MemberSummary{
		ID:         member.Id,
		Username:   member.Name,
		Rank:       member.Rank.String(),
		Attendance: attendanceCount,
		RSIHandle:  rsiHandle,
	}
	if balances != nil && len(balances) > 0 {
		if tokenBalance, ok := balances[member.Id]; ok {
			summary.TokenBalance = tokenBalance
		}
	}

	return summary
}

func paginateAttendanceRecords(records []AttendanceRecord, limit, page int) []AttendanceRecord {
	start := (page - 1) * limit
	if start >= len(records) {
		return []AttendanceRecord{}
	}

	end := min(start+limit, len(records))

	return records[start:end]
}

func paginateTokenTransactions(records []TokenTransaction, limit, page int) []TokenTransaction {
	start := (page - 1) * limit
	if start >= len(records) {
		return []TokenTransaction{}
	}

	end := min(start+limit, len(records))

	return records[start:end]
}

func normalizeSearch(search string) string {
	return strings.TrimSpace(strings.ToLower(search))
}

func fuzzyContains(haystack, needle string) bool {
	normalizedHaystack := normalizeSearch(haystack)
	normalizedNeedle := normalizeSearch(needle)
	needleRunes := []rune(normalizedNeedle)

	if len(needleRunes) == 0 {
		return true
	}

	if strings.Contains(normalizedHaystack, normalizedNeedle) {
		return true
	}

	needleIndex := 0
	for _, haystackChar := range normalizedHaystack {
		if haystackChar == needleRunes[needleIndex] {
			needleIndex += 1
			if needleIndex == len(needleRunes) {
				return true
			}
		}
	}

	return false
}

func matchesAnyField(search string, fields ...string) bool {
	if search == "" {
		return true
	}

	for _, field := range fields {
		if fuzzyContains(field, search) {
			return true
		}
	}

	return false
}
