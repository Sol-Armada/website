package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"strconv"
	"time"

	"github.com/pkg/errors"
	"github.com/sol-armada/admin/users"
	solmembers "github.com/sol-armada/sol-bot/members"
	"github.com/sol-armada/sol-bot/ranks"
	"go.mongodb.org/mongo-driver/mongo"
)

var membersActions = map[string]Action{
	"list":      getMembers,
	"me":        getMe,
	"update-me": updateMe,
}

type MembersCollection struct {
	*mongo.Collection
}

func getMe(ctx context.Context, c *Client, token any) CommandResponse {
	cr := CommandResponse{
		Thing:  "members",
		Action: "me",
	}

	uAccess, ok := ctx.Value(contextKeyAccess).(userAccess)
	if !ok {
		cr.Error = "unauthorized"
		return cr
	}

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

func updateMe(ctx context.Context, c *Client, arg any) CommandResponse {
	logger := slog.Default()

	cr := CommandResponse{
		Thing:  "members",
		Action: "update-me",
	}

	me := ctx.Value(contextKeyMember).(*solmembers.Member)
	_ = me
	updatesMap := map[string]interface{}{}
	if err := json.Unmarshal([]byte(arg.(string)), &updatesMap); err != nil {
		logger.Error("failed to parse me", "error", err)
		cr.Error = "internal_error"
		return cr
	}

	logger.Debug("updating me")

	me.Name = updatesMap["name"].(string)
	me.Age = int(updatesMap["age"].(float64))
	me.Playtime = int(updatesMap["playtime"].(float64))
	gameplayRaw := updatesMap["gameplay"].([]interface{})
	gameplayList := []solmembers.GameplayTypes{}
	for _, gameplay := range gameplayRaw {
		gameplayList = append(gameplayList, solmembers.GameplayTypes(gameplay.(string)))
	}
	me.Gameplay = gameplayList
	onboardedAtRaw := updatesMap["onboarded_at"].(string)
	onboardedAt, err := time.Parse(time.RFC3339, onboardedAtRaw)
	if err != nil {
		logger.Error("failed to parse onboarded_at", "error", err)
		cr.Error = "internal_error"
		return cr
	}
	me.OnboardedAt = &onboardedAt

	if err := me.Save(); err != nil {
		logger.Error("failed to save me", "error", err)
		cr.Error = "internal_error"
		return cr
	}

	return cr
}
