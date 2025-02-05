package main

import (
	"context"
	"log/slog"
	"strconv"

	"github.com/pkg/errors"
	"github.com/sol-armada/admin/users"
	attndnc "github.com/sol-armada/sol-bot/attendance"
	"github.com/sol-armada/sol-bot/members"
	"github.com/sol-armada/sol-bot/ranks"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var attendanceActions = map[string]Action{
	"list":    listAttendance,
	"count":   getAttendanceCount,
	"records": getMemberAttendanceRecords,
}

func listAttendance(ctx context.Context, _ *Client, arg any) CommandResponse {
	logger := slog.Default()

	member := ctx.Value(contextKeyMember).(*members.Member)

	cr := CommandResponse{
		Thing:  "attendance",
		Action: "list",
	}

	if arg == "undefined" {
		arg = "0"
	}

	if member.Rank > ranks.Lieutenant {
		cr.Error = "unauthorized"
		return cr
	}

	page, err := strconv.Atoi(arg.(string))
	if err != nil {
		logger.Error("failed to parse page", "error", err)
		cr.Error = "internal_error"
		return cr
	}

	if page < 1 {
		page = 1
	}

	attendanceRecords, err := attndnc.List(bson.D{}, 100, page)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			cr.Result = []*users.User{}
			return cr
		}

		logger.Error("failed to list attendance", "error", err)
		cr.Error = "internal_error"
		return cr
	}

	cr.Result = attendanceRecords

	return cr
}

func getAttendanceCount(_ context.Context, _ *Client, id any) CommandResponse {
	logger := slog.Default()

	memberId := id.(string)

	cr := CommandResponse{
		Thing:  "attendance",
		Action: "count",
	}

	count, err := attndnc.GetMemberAttendanceCount(memberId)
	if err != nil {
		logger.Error("failed to get attendance count", "error", err)
		cr.Error = "internal_error"
		return cr
	}

	cr.Result = count
	return cr
}

func getMemberAttendanceRecords(_ context.Context, _ *Client, id any) CommandResponse {
	logger := slog.Default()

	memberId := id.(string)

	cr := CommandResponse{
		Thing:  "attendance",
		Action: "records",
	}

	records, err := attndnc.GetMemberAttendanceRecords(memberId)
	if err != nil {
		logger.Error("failed to get attendance records", "error", err)
		cr.Error = "internal_error"
		return cr
	}

	cr.Result = records
	return cr
}
