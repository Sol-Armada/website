package main

import (
	"context"
	"log/slog"
	"time"

	"github.com/pkg/errors"
	"github.com/sol-armada/sol-bot/members"
	"github.com/sol-armada/sol-bot/ranks"
	tkns "github.com/sol-armada/sol-bot/tokens"
	"go.mongodb.org/mongo-driver/mongo"
)

var tokensActions = map[string]Action{
	"list": listTokenRecords,
}

func watchForTokens(ctx context.Context, hub *Hub) {
	logger := slog.Default()

	tokenRecord := make(chan tkns.TokenRecord)

	go func() {
		if err := tkns.Watch(ctx, tokenRecord); err != nil {
			logger.Error("failed to watch for token records", "error", err)
		}
	}()

	for {
		select {
		case <-ctx.Done():
			close(tokenRecord)
			return
		case tr := <-tokenRecord:
			logger.Debug("token record received", "record", tr)

			// iterate over hub clients
			for c, m := range hub.clients {
				if m == nil || !m.IsOfficer() {
					continue
				}

				res := CommandResponse{
					Thing:  "tokens",
					Action: "get",
					Result: tr,
				}
				c.send <- res.ToJsonBytes()
			}

			time.Sleep(1 * time.Second)
		}
	}
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
			cr.Result = []*members.Member{}
			return cr
		}

		logger.Error("failed to list token records", "error", err)
		cr.Error = "internal_error"
		return cr
	}

	cr.Result = tokenRecords

	return cr
}
