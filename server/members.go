package main

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"github.com/pkg/errors"
	"github.com/sol-armada/admin/ranks"
	"github.com/sol-armada/admin/users"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var membersActions = map[string]Action{
	"list": getMembers,
	"me":   getMe,
}

type membersCollection struct {
	*mongo.Collection
}

var client *mongo.Client
var members *membersCollection

func (m *membersCollection) GetMemberById(ctx context.Context, id string) (*users.User, error) {
	member := &users.User{}
	if err := m.FindOne(ctx, bson.M{"_id": id}).Decode(member); err != nil {
		return nil, err
	}

	return member, nil
}

func (m *membersCollection) GetMembers(ctx context.Context, page int) ([]*users.User, error) {
	var members []*users.User

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	opts := options.Find().SetSort(bson.M{"rank": 1}).SetLimit(100).SetSkip(int64(100 * (page - 1)))

	cursor, err := m.Find(ctx, bson.D{{Key: "is_bot", Value: bson.D{{Key: "$eq", Value: false}}}}, opts)
	if err != nil {
		return nil, err
	}
	if err := cursor.All(ctx, &members); err != nil {
		return nil, err
	}

	if len(members) == 0 {
		return nil, mongo.ErrNoDocuments
	}

	return members, nil
}

func (m *membersCollection) UpdateMember(ctx context.Context, member *users.User) error {
	_, err := m.UpdateOne(ctx, bson.M{"_id": member.ID}, bson.M{"$set": member})
	return err
}

func setupMembersStore(ctx context.Context, host string, port int, username string, password string, database string) error {
	if client != nil {
		return nil
	}

	usernamePassword := username + ":" + password + "@"
	if usernamePassword == ":@" {
		usernamePassword = ""
	}

	uri := fmt.Sprintf("mongodb://%s%s:%d", usernamePassword, host, port)
	clientOptions := options.Client().ApplyURI(uri).SetConnectTimeout(5 * time.Second)
	c, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return errors.Wrap(err, "creating new store")
	}
	client = c

	members = &membersCollection{client.Database(database).Collection("users")}

	return nil
}

func getMe(ctx context.Context, c *Client, token any) CommandResponse {
	cr := CommandResponse{
		Thing:  "members",
		Action: "me",
	}

	uAccess := ctx.Value(contextKeyAccess).(userAccess)

	logger := slog.With("token", uAccess.Token)
	logger.Info("creating new user access")

	user := &users.User{}

	userMap, err := getDiscordMe(uAccess)
	if err != nil {
		if err.Error() != "invalid_grant" {
			logger.Error("failed to get user", "error", err)
			cr.Error = "internal_error"
		}

		cr.Error = err.Error()
		return cr
	}

	user.ID = userMap["id"].(string)

	user, err = members.GetMemberById(ctx, user.ID)
	if err != nil {
		if !errors.Is(err, mongo.ErrNoDocuments) {
			logger.Error("failed to get user", "error", err)
			cr.Error = "internal_error"
			return cr
		}

		user.ID = userMap["id"].(string)
		user.Name = userMap["username"].(string)
		user.Rank = ranks.Guest
		avatar, ok := userMap["avatar"].(string)
		if !ok {
			avatar = ""
		}
		user.Avatar = avatar

		cr.Result = user
		return cr
	}

	if user.Avatar == "" {
		user.Avatar = userMap["avatar"].(string)

		if err := members.UpdateMember(ctx, user); err != nil {
			logger.Error("failed to update user", "error", err)
			cr.Error = "internal_error"
			return cr
		}
	}

	cr.Result = user

	return cr
}

func getMembers(ctx context.Context, c *Client, arg any) CommandResponse {

	logger := slog.Default()

	user := ctx.Value(contextKeyMember).(*users.User)

	cr := CommandResponse{
		Thing:  "members",
		Action: "list",
	}

	if arg == "undefined" {
		cr.Result = []*users.User{}
		return cr
	}

	if user.Rank > ranks.Lieutenant {
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

	m, err := members.GetMembers(ctx, page)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			cr.Result = []*users.User{}
			return cr
		}
		logger.Error("failed to list users", "error", err)
		cr.Error = "internal_error"
		return cr
	}

	cr.Result = m

	return cr
}
