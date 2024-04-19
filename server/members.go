package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"github.com/sol-armada/admin/ranks"
	"github.com/sol-armada/admin/stores"
	"github.com/sol-armada/admin/users"
	"github.com/spf13/viper"
	"go.mongodb.org/mongo-driver/mongo"
)

var membersActions = map[string]Action{
	"get": getMembers,
	"me":  getMe,
}

func getMe(c *Client, arg any) {
	host := viper.GetString("MONGO.HOST")
	port := viper.GetInt("MONGO.PORT")
	database := viper.GetString("MONGO.DATABASE")
	if err := stores.Setup(context.Background(), host, port, "", "", database); err != nil {
		slog.Error("failed to setup stores", "error", err)
		return
	}

	cr := CommandResponse{
		Thing: "me",
	}
	defer func() {
		j, _ := json.Marshal(cr)
		c.send <- j
	}()

	token := arg.(string)

	logger := slog.With("token", token)
	logger.Info("creating new user access")

	redirectURI := viper.GetString("DISCORD.REDIRECT_URI")
	clientId := viper.GetString("DISCORD.CLIENT_ID")
	clientSecret := viper.GetString("DISCORD.CLIENT_SECRET")

	logger = logger.With("redirect_uri", redirectURI, "client_id", clientId, "client_secret", clientSecret)

	logger.Debug("sending auth request")

	req, err := http.NewRequest("GET", "https://discord.com/api/v10/oauth2/@me", nil)
	if err != nil {
		logger.Error("failed to create request", "error", err)
		cr.Error = err.Error()
		return
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	logger.Debug("req for @me to discord")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		logger.Error("failed to send request", "error", err)
		cr.Error = err.Error()
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		logger.Error("unauthorized", "status", resp.StatusCode)
		cr.Error = "invalid_grant"
		return
	}

	if resp.StatusCode == http.StatusBadRequest {
		errorMessage, _ := io.ReadAll(resp.Body)
		type ErrorMessage struct {
			ErrorType   string `json:"error"`
			Description string `json:"error_description"`
		}
		errMsg := ErrorMessage{}
		if err := json.Unmarshal(errorMessage, &errMsg); err != nil {
			cr.Error = err.Error()
			return
		}
		cr.Error = errMsg.Description
		return
	}

	res := map[string]interface{}{}
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		logger.Error("failed to decode access", "error", err)
		cr.Error = err.Error()
		return
	}
	userMap := res["user"].(map[string]interface{})

	user := &users.User{}

	user.ID = userMap["id"].(string)

	if err := stores.Users.Get(user.ID).Decode(user); err != nil {
		if !errors.Is(err, mongo.ErrNoDocuments) {
			logger.Error("failed to get user", "error", err)
			cr.Error = err.Error()
			return
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
		return
	}

	if user.Avatar == "" {
		user.Avatar = userMap["avatar"].(string)

		if err := stores.Users.Update(user.ID, user); err != nil {
			logger.Error("failed to update user", "error", err)
			cr.Error = err.Error()
			return
		}
	}

	cr.Result = user
}

func getMembers(c *Client, arg any) {
	members := []*users.User{
		{
			ID:   "1",
			Name: "KooTheGreat",
			Rank: ranks.Admiral,
		},
		{
			ID:   "2",
			Name: "KooTheGreat2",
			Rank: ranks.Admiral,
		},
	}

	cr := CommandResponse{
		Thing:  "members",
		Result: members,
	}

	j, _ := json.Marshal(cr)

	c.send <- j
}
