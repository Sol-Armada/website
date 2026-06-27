package service

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"slices"
	"sort"
	"strconv"
	"strings"
	"time"

	"log/slog"

	"github.com/sol-armada/sol-bot/attendance"
	"github.com/sol-armada/sol-bot/members"
	"github.com/sol-armada/sol-bot/tokens"
)

var (
	ErrAttendanceRecordNotFound = errors.New("attendance record not found")
	ErrInvalidAttendanceInput   = errors.New("invalid attendance input")
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
	AwardTokens      bool      `json:"awardTokens"`
	DateCreated      time.Time `json:"dateCreated"`
}

type TokenTransaction struct {
	ID             string    `json:"id"`
	MemberID       string    `json:"memberId"`
	MemberName     string    `json:"memberName,omitempty"`
	Amount         int       `json:"amount"`
	Reason         string    `json:"reason"`
	CreatedAt      time.Time `json:"createdAt"`
	Comment        string    `json:"comment,omitempty"`
	GiverID        string    `json:"giverId,omitempty"`
	AttendanceID   string    `json:"attendanceId,omitempty"`
	AttendanceName string    `json:"attendanceName,omitempty"`
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
	ProfileImage string `json:"profileImage,omitempty"`
	OnTime       *bool  `json:"onTime,omitempty"`
	IsManager    *bool  `json:"isManager,omitempty"`
}

type AttendanceEditPayload struct {
	Record       AttendanceRecord `json:"record"`
	Participants []MemberSummary  `json:"participants"`
}

type UpdateAttendanceRecordInput struct {
	Name                 string   `json:"name"`
	Recorded             bool     `json:"recorded"`
	Successful           bool     `json:"successful"`
	AwardTokens          bool     `json:"awardTokens"`
	ParticipantIds       []string `json:"participantIds"`
	OnTimeParticipantIds []string `json:"onTimeParticipantIds"`
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
			AwardTokens:      att.Tokenable,
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

	memberMap := make(map[string]*members.Member)
	for _, token := range allTokens {
		memberMap[token.MemberId] = nil
		if token.GiverId != nil {
			memberMap[*token.GiverId] = nil
		}
	}

	members, err := members.GetList(slices.Collect(maps.Keys(memberMap)))
	if err != nil {
		return nil, fmt.Errorf("failed to fetch member details: %w", err)
	}

	for _, m := range members {
		memberMap[m.Id] = m
	}

	attendanceMap := make(map[string]*attendance.Attendance)
	for _, token := range allTokens {
		if token.AttendanceId != nil {
			attendanceMap[*token.AttendanceId] = nil
		}
	}

	attendances, err := attendance.ListByIds(slices.Collect(maps.Keys(attendanceMap)))
	if err != nil {
		return nil, fmt.Errorf("failed to fetch attendance details: %w", err)
	}

	for _, a := range attendances {
		attendanceMap[a.Id] = a
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
				func() string {
					if member, ok := memberMap[t.MemberId]; ok && member != nil {
						return member.Name
					}
					return ""
				}(),
				strconv.Itoa(t.Amount),
				string(t.Reason),
				comment,
				func() string {
					if giver, ok := memberMap[giverId]; ok && giver != nil {
						return giver.Name
					}
					return ""
				}(),
				func() string {
					if att, ok := attendanceMap[attendanceId]; ok && att != nil {
						return att.Name
					}
					return ""
				}(),
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
			AttendanceName: func() string {
				if att, ok := attendanceMap[attendanceId]; ok && att != nil {
					return att.Name
				}
				return ""
			}(),
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

type CreateAttendanceRecordInput struct {
	SubmittedBy    string   `json:"submittedBy"`
	Name           string   `json:"name"`
	ParticipantIds []string `json:"participantIds"`
	ManagerIds     []string `json:"managerIds"`
	AwardTokens    bool     `json:"awardTokens"`
}

func (s *AdminService) CreateAttendanceRecord(ctx context.Context, input CreateAttendanceRecordInput) error {
	allMembers := make(map[string]*members.Member)
	for _, id := range input.ParticipantIds {
		allMembers[id] = nil
	}
	for _, id := range input.ManagerIds {
		allMembers[id] = nil
	}
	if input.SubmittedBy != "" {
		allMembers[input.SubmittedBy] = nil
	}

	memberIDs := make([]string, 0, len(allMembers))
	for id := range allMembers {
		memberIDs = append(memberIDs, id)
	}

	membersList, err := members.GetList(memberIDs)
	if err != nil {
		return fmt.Errorf("failed to fetch member details: %w", err)
	}

	for _, m := range membersList {
		allMembers[m.Id] = m
	}

	newAttendance, err := attendance.New(input.Name, allMembers[input.SubmittedBy])
	if err != nil {
		return fmt.Errorf("failed to create new attendance record: %w", err)
	}

	for _, participantId := range input.ParticipantIds {
		if member, ok := allMembers[participantId]; ok && member != nil {
			err = newAttendance.AddParticipant(member)
			if err != nil {
				s.logger.Error("Failed to add participant to attendance record", "memberId", participantId, "error", err)
			}
			continue
		}
		s.logger.Warn("Participant ID not found in members list", "memberId", participantId)
	}

	for _, managerId := range input.ManagerIds {
		if member, ok := allMembers[managerId]; ok && member != nil {
			err = newAttendance.SetParticipantManager(managerId)
			if err != nil {
				s.logger.Error("Failed to add manager to attendance record", "memberId", managerId, "error", err)
			}
			continue
		}
		s.logger.Warn("Manager ID not found in members list", "memberId", managerId)
	}

	return nil
}

func (s *AdminService) GetAttendanceRecord(ctx context.Context, attendanceID string) (*AttendanceRecord, error) {
	if strings.TrimSpace(attendanceID) == "" {
		return nil, nil
	}

	att, err := attendance.Get(attendanceID)
	if err != nil {
		if errors.Is(err, attendance.ErrAttendanceNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to fetch attendance record %s: %w", attendanceID, err)
	}

	submittedBy := ""
	if att.SubmittedBy != nil {
		submittedBy = att.SubmittedBy.Name
	}

	participants, err := att.Participants()
	participantCount := 0
	if err == nil {
		participantCount = len(participants)
	}

	record := &AttendanceRecord{
		ID:               att.Id,
		Name:             att.Name,
		SubmittedBy:      submittedBy,
		ParticipantCount: participantCount,
		Recorded:         att.Recorded,
		Successful:       att.Successful,
		AwardTokens:      att.Tokenable,
		DateCreated:      att.DateCreated,
	}

	return record, nil
}

func (s *AdminService) GetMembersByAttendance(ctx context.Context, attendanceID string) ([]MemberSummary, error) {
	if strings.TrimSpace(attendanceID) == "" {
		return nil, nil
	}

	att, err := attendance.Get(attendanceID)
	if err != nil {
		if errors.Is(err, attendance.ErrAttendanceNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to fetch attendance record %s: %w", attendanceID, err)
	}

	participants, err := att.Participants()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch participants for attendance record %s: %w", attendanceID, err)
	}

	result := make([]MemberSummary, 0, len(participants))
	for _, p := range participants {
		if p.Member == nil {
			continue
		}

		summary := buildMemberSummary(*p.Member, nil)
		onTime := p.JoinedAtStart
		isManager := p.IsManager
		summary.OnTime = &onTime
		summary.IsManager = &isManager
		result = append(result, summary)
	}

	return result, nil
}

func (s *AdminService) GetAttendanceEditPayload(ctx context.Context, attendanceID string) (*AttendanceEditPayload, error) {
	record, err := s.GetAttendanceRecord(ctx, attendanceID)
	if err != nil {
		return nil, err
	}
	if record == nil {
		return nil, nil
	}

	participants, err := s.GetMembersByAttendance(ctx, attendanceID)
	if err != nil {
		return nil, err
	}

	return &AttendanceEditPayload{
		Record:       *record,
		Participants: participants,
	}, nil
}

func (s *AdminService) UpdateAttendanceRecord(ctx context.Context, attendanceID string, input UpdateAttendanceRecordInput) error {
	if strings.TrimSpace(attendanceID) == "" {
		return fmt.Errorf("%w: attendance ID is required", ErrInvalidAttendanceInput)
	}

	name := strings.TrimSpace(input.Name)
	if name == "" {
		return fmt.Errorf("%w: event name is required", ErrInvalidAttendanceInput)
	}

	if input.Successful && !input.Recorded {
		return fmt.Errorf("%w: successful attendance must be recorded", ErrInvalidAttendanceInput)
	}

	participantIDs := uniqueNonEmptyStrings(input.ParticipantIds)
	if len(participantIDs) == 0 {
		return fmt.Errorf("%w: at least one participant is required", ErrInvalidAttendanceInput)
	}

	onTimeIDs := uniqueNonEmptyStrings(input.OnTimeParticipantIds)
	onTimeSet := make(map[string]struct{}, len(onTimeIDs))
	for _, id := range onTimeIDs {
		onTimeSet[id] = struct{}{}
	}

	participantSet := make(map[string]struct{}, len(participantIDs))
	for _, id := range participantIDs {
		participantSet[id] = struct{}{}
	}

	for _, id := range onTimeIDs {
		if _, ok := participantSet[id]; !ok {
			return fmt.Errorf("%w: on-time participant %s is not in participant list", ErrInvalidAttendanceInput, id)
		}
	}

	att, err := attendance.Get(attendanceID)
	if err != nil {
		if errors.Is(err, attendance.ErrAttendanceNotFound) {
			return fmt.Errorf("%w: %s", ErrAttendanceRecordNotFound, attendanceID)
		}

		return fmt.Errorf("failed to fetch attendance record %s: %w", attendanceID, err)
	}

	if att.Name != name {
		att.Name = name
	}

	att.Recorded = input.Recorded
	att.Successful = input.Successful
	att.Tokenable = input.AwardTokens
	if input.Recorded {
		att.Status = attendance.AttendanceStatusRecorded
	} else if att.Status == attendance.AttendanceStatusRecorded {
		att.Status = attendance.AttendanceStatusActive
	}

	currentParticipants, err := att.Participants()
	if err != nil {
		return fmt.Errorf("failed to fetch attendance participants: %w", err)
	}

	currentByID := make(map[string]*members.Member, len(currentParticipants))
	for _, participant := range currentParticipants {
		if participant == nil || participant.Member == nil {
			continue
		}

		currentByID[participant.Member.Id] = participant.Member
	}

	for id, member := range currentByID {
		if _, keep := participantSet[id]; keep {
			continue
		}

		if err := att.RemoveParticipant(member); err != nil {
			return fmt.Errorf("failed to remove participant %s: %w", id, err)
		}
	}

	toAddIDs := make([]string, 0)
	for _, id := range participantIDs {
		if _, exists := currentByID[id]; exists {
			continue
		}

		toAddIDs = append(toAddIDs, id)
	}

	if len(toAddIDs) > 0 {
		memberList, err := members.GetList(toAddIDs)
		if err != nil {
			return fmt.Errorf("failed to fetch participants to add: %w", err)
		}

		membersByID := make(map[string]*members.Member, len(memberList))
		for _, member := range memberList {
			if member == nil {
				continue
			}

			membersByID[member.Id] = member
		}

		missingParticipantIDs := make([]string, 0)
		for _, id := range toAddIDs {
			member, ok := membersByID[id]
			if !ok || member == nil {
				missingParticipantIDs = append(missingParticipantIDs, id)
				continue
			}

			if err := att.AddParticipant(member); err != nil {
				return fmt.Errorf("failed to add participant %s: %w", id, err)
			}
		}

		if len(missingParticipantIDs) > 0 {
			return fmt.Errorf("%w: participant IDs not found: %s", ErrInvalidAttendanceInput, strings.Join(missingParticipantIDs, ", "))
		}
	}

	updatedParticipants, err := att.Participants()
	if err != nil {
		return fmt.Errorf("failed to fetch updated participants: %w", err)
	}

	for _, participant := range updatedParticipants {
		if participant == nil || participant.Member == nil {
			continue
		}

		_, shouldBeOnTime := onTimeSet[participant.Member.Id]
		if participant.JoinedAtStart == shouldBeOnTime {
			continue
		}

		if err := participant.SetJoinedAtStart(att.Id, shouldBeOnTime); err != nil {
			return fmt.Errorf("failed to update on-time flag for participant %s: %w", participant.Member.Id, err)
		}
	}

	if err := att.Save(); err != nil {
		return fmt.Errorf("failed to save attendance record: %w", err)
	}

	return nil
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
		ID:           member.Id,
		Username:     member.Name,
		Rank:         member.Rank.String(),
		Attendance:   attendanceCount,
		RSIHandle:    rsiHandle,
		ProfileImage: member.Avatar,
	}
	if len(balances) > 0 {
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
		if field == "" {
			continue
		}

		if fuzzyContains(field, search) {
			return true
		}
	}

	return false
}

func uniqueNonEmptyStrings(values []string) []string {
	if len(values) == 0 {
		return []string{}
	}

	seen := make(map[string]struct{}, len(values))
	result := make([]string, 0, len(values))

	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}

		if _, exists := seen[trimmed]; exists {
			continue
		}

		seen[trimmed] = struct{}{}
		result = append(result, trimmed)
	}

	return result
}
