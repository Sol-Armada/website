package main

import (
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
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
}

var loginActions = map[string]Action{
	"auth":    authenticate,
	"refresh": refresh,
}

func authenticate(c *Client, arg any) {
	cr := CommandResponse{}
	cr.Thing = "login"
	defer func() {
		j, _ := json.Marshal(cr)
		c.send <- j
	}()

	code := arg.(string)

	logger := slog.With("code", code)
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

	userAccess := userAccess{
		AccessToken:  userAccessMap["access_token"].(string),
		RefreshToken: userAccessMap["refresh_token"].(string),
		ExpiresAt:    expiresAt,
	}

	logger.Debug("created new user access", "access", userAccess)

	cr.Result = userAccess
}

func refresh(c *Client, arg any) {
	cr := CommandResponse{}
	cr.Thing = "login"
	defer func() {
		j, _ := json.Marshal(cr)
		c.send <- j
	}()

	token := arg.(string)

	logger := slog.With("token", token)
	logger.Info("refreshing access token")

	clientId := viper.GetString("DISCORD.CLIENT_ID")
	clientSecret := viper.GetString("DISCORD.CLIENT_SECRET")

	logger = logger.With("client_id", clientId, "client_secret", clientSecret)

	data := url.Values{}
	data.Set("client_id", clientId)
	data.Set("client_secret", clientSecret)
	data.Set("refresh_token", token)
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
	expiresAt := time.Now().Add(time.Second * time.Duration(userAccessMap["expires_in"].(int)))

	userAccess := userAccess{
		AccessToken:  userAccessMap["access_token"].(string),
		RefreshToken: userAccessMap["refresh_token"].(string),
		ExpiresAt:    expiresAt,
	}

	logger.Debug("created new user access", "access", userAccess)

	cr.Result = userAccess

}
