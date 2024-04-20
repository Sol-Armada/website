package main

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type userAccess struct {
	Token        string    `json:"token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
	Id           string    `json:"id"`
}

var loginActions = map[string]Action{
	"auth":    authenticate,
	"refresh": refresh,
}

func authenticate(ctx context.Context, c *Client, arg any) {
	cr := CommandResponse{
		Thing:  "login",
		Action: "auth",
	}
	defer func() {
		j, _ := json.Marshal(cr)
		c.send <- j
	}()

	code := arg.(string)

	logger := slog.Default().With("code", code)
	logger.Info("creating new user access")

	redirectURI := viper.GetString("DISCORD.REDIRECT_URI")
	clientId := viper.GetString("DISCORD.CLIENT_ID")
	clientSecret := viper.GetString("DISCORD.CLIENT_SECRET")

	logger = logger.With("redirect_uri", redirectURI, "client_id", clientId, "client_secret", clientSecret)

	data := url.Values{}
	data.Set("client_id", clientId)
	data.Set("client_secret", clientSecret)
	data.Set("redirect_uri", redirectURI)
	data.Set("grant_type", "authorization_code")
	data.Set("code", code)

	logger.Debug("sending auth request")

	req, err := http.NewRequest("POST", "https://discord.com/api/v10/oauth2/token", strings.NewReader(data.Encode()))
	if err != nil {
		logger.Error("failed to create request", "error", err)
		cr.Error = err.Error()
		return
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	logger.Debug("req for authentication to discord")

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
		if errMsg.ErrorType == "invalid_grant" {
			cr.Error = "invalid_grant"
			return
		}

		cr.Error = errMsg.Description
		return
	}

	access := map[string]interface{}{}
	if err := json.NewDecoder(resp.Body).Decode(&access); err != nil {
		logger.Error("failed to decode access", "error", err)
		cr.Error = err.Error()
		return
	}

	userAccessMap := map[string]interface{}{}
	uaj, _ := json.Marshal(access)
	if err := json.Unmarshal(uaj, &userAccessMap); err != nil {
		logger.Error("failed to unmarshal user access", "error", err)
		cr.Error = err.Error()
		return
	}

	// convert expires in from int to time
	expiresAt := time.Now().Add(time.Second * time.Duration(userAccessMap["expires_in"].(float64))).UTC()

	uAccess := userAccess{
		Token:        userAccessMap["access_token"].(string),
		RefreshToken: userAccessMap["refresh_token"].(string),
		ExpiresAt:    expiresAt,
	}

	logger.Debug("created new user access", "access", uAccess)

	j, _ := json.Marshal(uAccess)
	ecyrptedAccess, err := encrypt(string(j))
	if err != nil {
		logger.Error("failed to encrypt user access", "error", err)
		cr.Error = err.Error()
		return
	}

	cr.Result = ecyrptedAccess
}

func refresh(ctx context.Context, c *Client, arg any) {
	cr := CommandResponse{
		Thing:  "login",
		Action: "refresh",
	}
	defer func() {
		j, _ := json.Marshal(cr)
		c.send <- j
	}()

	uAccess := ctx.Value(contextKeyAccess).(userAccess)

	logger := slog.With("refresh_token", uAccess.RefreshToken)
	logger.Info("refreshing access token")

	clientId := viper.GetString("DISCORD.CLIENT_ID")
	clientSecret := viper.GetString("DISCORD.CLIENT_SECRET")

	logger = logger.With("client_id", clientId, "client_secret", clientSecret)

	data := url.Values{}
	data.Set("client_id", clientId)
	data.Set("client_secret", clientSecret)
	data.Set("refresh_token", uAccess.RefreshToken)
	data.Set("grant_type", "refresh_token")

	logger.Debug("sending auth request")

	req, err := http.NewRequest("POST", "https://discord.com/api/v10/oauth2/token", strings.NewReader(data.Encode()))
	if err != nil {
		logger.Error("failed to create request", "error", err)
		cr.Error = err.Error()
		return
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	logger.Debug("req for authentication to discord")

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

	discordAccessMap := map[string]interface{}{}
	if err := json.NewDecoder(resp.Body).Decode(&discordAccessMap); err != nil {
		logger.Error("failed to decode access", "error", err)
		cr.Error = err.Error()
		return
	}

	// convert expires in from int to time
	expiresAt := time.Now().Add(time.Second * time.Duration(discordAccessMap["expires_in"].(int)))

	uAccess = userAccess{
		Token:        discordAccessMap["access_token"].(string),
		RefreshToken: discordAccessMap["refresh_token"].(string),
		ExpiresAt:    expiresAt,
	}

	logger.Debug("created new user access", "access", uAccess)

	j, _ := json.Marshal(uAccess)
	ecyrptedAccess, err := encrypt(string(j))
	if err != nil {
		logger.Error("failed to encrypt user access", "error", err)
		cr.Error = err.Error()
		return
	}

	cr.Result = ecyrptedAccess
}
