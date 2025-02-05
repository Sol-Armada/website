package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"strconv"
	"time"

	"github.com/pkg/errors"
	attndnc "github.com/sol-armada/sol-bot/attendance"
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

		logger.Debug("member not found")

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

	members, err := solmembers.List(page)
	if err != nil {
		logger.Error("failed to list users", "error", err)
		cr.Error = "internal_error"
		return cr
	}

	// get attendance count
	memberCounts := map[string]int{}
	for _, member := range members {
		count, err := attndnc.GetMemberAttendanceCount(member.Id)
		if err != nil {
			logger.Error("failed to get attendance count", "error", err)
			cr.Error = "internal_error"
			return cr
		}

		memberCounts[member.Id] = count
	}

	cr.Result = map[string]interface{}{
		"members":      members,
		"event_counts": memberCounts,
	}

	return cr
}

func updateMe(ctx context.Context, c *Client, arg any) CommandResponse {
	logger := slog.Default()

	cr := CommandResponse{
		Thing:  "members",
		Action: "update-me",
	}

	me := ctx.Value(contextKeyMember).(*solmembers.Member)

	updatesMap := map[string]interface{}{}
	if err := json.Unmarshal([]byte(arg.(string)), &updatesMap); err != nil {
		logger.Error("failed to parse me", "error", err)
		cr.Error = "internal_error"
		return cr
	}

	logger = logger.With("updates", updatesMap)

	logger.Debug("updating me")

	me.Name = updatesMap["name"].(string)

	// onboarding stuff
	me.Age = int(updatesMap["age"].(float64))
	me.Playtime = int(updatesMap["playTime"].(float64))
	if rawGameplayList, ok := updatesMap["gameplay"].([]interface{}); ok {
		gameplayList := []solmembers.GameplayType{}
		for _, rawGameplay := range rawGameplayList {
			gameplayList = append(gameplayList, solmembers.ToGameplayType(rawGameplay.(string)))
		}
		me.Gameplay = gameplayList
	}
	me.TimeZone = updatesMap["timeZone"].(string)
	me.FoundBy = updatesMap["foundBy"].(string)
	me.Other = ""
	me.Recruiter = nil

	switch me.FoundBy {
	case "other":
		me.Other = updatesMap["other"].(string)
	case "recruited":
		recruiter, err := solmembers.Get(updatesMap["recruitedBy"].(string))
		if err != nil {
			logger.Error("failed to get recruiter", "error", err)
			cr.Error = "internal_error"
			return cr
		}

		me.Recruiter = &recruiter.Id
	}

	onboardedAtRaw := updatesMap["onboardedAt"].(string)
	onboardedAt, err := time.Parse(time.RFC3339, onboardedAtRaw)
	if err != nil {
		logger.Error("failed to parse onboardedAt", "error", err)
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
