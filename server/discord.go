package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

func authenticateDiscord(code string) (map[string]interface{}, error) {
	redirectURI := viper.GetString("DISCORD.REDIRECT_URI")
	clientId := viper.GetString("DISCORD.CLIENT_ID")
	clientSecret := viper.GetString("DISCORD.CLIENT_SECRET")

	logger := slog.Default().With("redirect_uri", redirectURI, "client_id", clientId, "client_secret", clientSecret)

	data := url.Values{}
	data.Set("client_id", clientId)
	data.Set("client_secret", clientSecret)
	data.Set("redirect_uri", redirectURI)
	data.Set("grant_type", "authorization_code")
	data.Set("code", code)

	logger.Debug("sending auth request")

	req, err := http.NewRequest("POST", "https://discord.com/api/v10/oauth2/token", strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	logger.Debug("req for authentication to discord")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return nil, errors.New("invalid_grant")
	}

	if resp.StatusCode == http.StatusBadRequest {
		errorMessage, _ := io.ReadAll(resp.Body)
		type ErrorMessage struct {
			ErrorType   string `json:"error"`
			Description string `json:"error_description"`
		}
		errMsg := ErrorMessage{}
		if err := json.Unmarshal(errorMessage, &errMsg); err != nil {
			return nil, err
		}
		if errMsg.ErrorType == "invalid_grant" {
			return nil, errors.New("invalid_grant")
		}

		return nil, errors.New(errMsg.Description)
	}

	access := map[string]interface{}{}
	if err := json.NewDecoder(resp.Body).Decode(&access); err != nil {
		return nil, err
	}

	return access, nil
}

func getDiscordMe(access userAccess) (map[string]interface{}, error) {
	req, err := http.NewRequest("GET", "https://discord.com/api/v10/oauth2/@me", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", access.Token))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return nil, errors.New("invalid_grant")
	}

	if resp.StatusCode == http.StatusBadRequest {
		errorMessage, _ := io.ReadAll(resp.Body)
		type ErrorMessage struct {
			ErrorType   string `json:"error"`
			Description string `json:"error_description"`
		}
		errMsg := ErrorMessage{}
		if err := json.Unmarshal(errorMessage, &errMsg); err != nil {
			return nil, err
		}

		return nil, errors.New(errMsg.Description)
	}

	res := map[string]interface{}{}
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, err
	}

	return res["user"].(map[string]interface{}), nil
}
