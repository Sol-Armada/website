package main

import (
	"context"
	"log/slog"
	"strconv"

	"github.com/pkg/errors"
	"github.com/sol-armada/admin/users"
	solmembers "github.com/sol-armada/sol-bot/members"
	"github.com/sol-armada/sol-bot/ranks"
	"go.mongodb.org/mongo-driver/mongo"
)

var membersActions = map[string]Action{
	"list": getMembers,
	"me":   getMe,
}

type MembersCollection struct {
	*mongo.Collection
}

func getMe(ctx context.Context, c *Client, token any) CommandResponse {
	cr := CommandResponse{
		Thing:  "members",
		Action: "me",
	}

	uAccess := ctx.Value(contextKeyAccess).(userAccess)

	logger := slog.With("token", uAccess.Token)
	logger.Info("creating new user access")

	member := &solmembers.Member{}

	discordUserMap, err := getDiscordMe(uAccess)
	if err != nil {
		if err.Error() != "invalid_grant" {
			logger.Error("failed to get user", "error", err)
			cr.Error = "internal_error"
		}

		cr.Error = err.Error()
		return cr
	}

	member.Id = discordUserMap["id"].(string)

	member, err = solmembers.Get(member.Id)
	if err != nil {
		if !errors.Is(err, mongo.ErrNoDocuments) {
			logger.Error("failed to get user", "error", err)
			cr.Error = "internal_error"
			return cr
		}

		member.Id = discordUserMap["id"].(string)
		member.Name = discordUserMap["username"].(string)
		member.Rank = ranks.None
		avatar, ok := discordUserMap["avatar"].(string)
		if !ok {
			avatar = ""
		}
		member.Avatar = avatar

		cr.Result = member
		return cr
	}

	if member.Avatar == "" {
		member.Avatar = discordUserMap["avatar"].(string)

		if err := member.Save(); err != nil {
			logger.Error("failed to save member", "error", err)
			cr.Error = "internal_error"
			return cr
		}

		// if err := membersCollection.UpdateMember(ctx, user); err != nil {
		// 	logger.Error("failed to update user", "error", err)
		// 	cr.Error = "internal_error"
		// 	return cr
		// }
	}

	cr.Result = member

	return cr
}

func getMembers(ctx context.Context, c *Client, arg any) CommandResponse {

	logger := slog.Default()

	member := ctx.Value(contextKeyMember).(*solmembers.Member)

	cr := CommandResponse{
		Thing:  "members",
		Action: "list",
	}

	if arg == "undefined" {
		cr.Result = []*users.User{}
		return cr
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

	members, err := solmembers.List(page)
	if err != nil {
		logger.Error("failed to list users", "error", err)
		cr.Error = "internal_error"
		return cr
	}

	cr.Result = members

	return cr
}
