package main

import (
	"context"
	"log/slog"

	"github.com/pkg/errors"
	"github.com/sol-armada/admin/users"
	"github.com/sol-armada/sol-bot/members"
	"github.com/sol-armada/sol-bot/ranks"
	tkns "github.com/sol-armada/sol-bot/tokens"
	"go.mongodb.org/mongo-driver/mongo"
)

var tokensActions = map[string]Action{
	"list": listTokenRecords,
}

func listTokenRecords(ctx context.Context, c *Client, arg any) CommandResponse {
	logger := slog.Default()

	member := ctx.Value(contextKeyMember).(*members.Member)

	cr := CommandResponse{
		Thing:  "tokens",
		Action: "list",
	}

	if member.Rank > ranks.Lieutenant {
		cr.Error = "unauthorized"
		return cr
	}

	tokenRecords, err := tkns.GetAll()
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			cr.Result = []*users.User{}
			return cr
		}

		logger.Error("failed to list token records", "error", err)
		cr.Error = "internal_error"
		return cr
	}

	cr.Result = tokenRecords

	return cr
}
