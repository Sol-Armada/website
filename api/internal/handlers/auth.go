package handlers

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
	"golang.org/x/oauth2"

	"github.com/sol-armada/website/internal/auth"
	"github.com/sol-armada/website/internal/dto"
	"github.com/sol-armada/website/internal/service"
)

const (
	DiscordAPIURL      = "https://discord.com/api/v10"
	DiscordAuthURL     = "https://discord.com/api/oauth2/authorize"
	DiscordTokenURL    = "https://discord.com/api/oauth2/token"
)

// AuthHandler handles authentication endpoints
type AuthHandler struct {
	oauthConfig     *oauth2.Config
	tokenService    *auth.TokenService
	cookieService   *auth.CookieService
	sessionService  *service.SessionService
	guildID         string // Required guild ID for access
	adminRoleID     string
	modRoleID       string
	log             *logrus.Logger
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(
	clientID, clientSecret, redirectURI string,
	scopes []string,
	tokenService *auth.TokenService,
	cookieService *auth.CookieService,
	sessionService *service.SessionService,
	guildID, adminRoleID, modRoleID string,
	log *logrus.Logger,
) *AuthHandler {
	return &AuthHandler{
		oauthConfig: &oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			RedirectURL:  redirectURI,
			Scopes:       scopes,
			Endpoint: oauth2.Endpoint{
				AuthURL:  DiscordAuthURL,
				TokenURL: DiscordTokenURL,
			},
		},
		tokenService:   tokenService,
		cookieService:  cookieService,
		sessionService: sessionService,
		guildID:        guildID,
		adminRoleID:    adminRoleID,
		modRoleID:      modRoleID,
		log:            log,
	}
}

// Login initiates the Discord OAuth flow
func (h *AuthHandler) Login(c echo.Context) error {
	// Generate random state for CSRF protection
	state := generateRandomState()
	
	// Determine if we should use secure cookies
	isProduction := h.cookieService != nil // Will be set properly in production

	// Store state in session cookie temporarily (5 minutes)
	stateCookie := &http.Cookie{
		Name:     "oauth_state",
		Value:    state,
		Path:     "/",
		MaxAge:   300, // 5 minutes
		HttpOnly: true,
		Secure:   isProduction,
		SameSite: http.SameSiteLaxMode,
	}
	c.SetCookie(stateCookie)

	// Redirect to Discord OAuth
	url := h.oauthConfig.AuthCodeURL(state, oauth2.AccessTypeOnline)
	return c.Redirect(http.StatusTemporaryRedirect, url)
}

// Callback handles the OAuth callback from Discord
func (h *AuthHandler) Callback(c echo.Context) error {
	// Verify state parameter
	stateCookie, err := c.Cookie("oauth_state")
	if err != nil {
		h.log.WithError(err).Error("Missing OAuth state cookie")
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "invalid_state",
			Message: "OAuth state mismatch",
		})
	}

	state := c.QueryParam("state")
	if state == "" || state != stateCookie.Value {
		h.log.Error("OAuth state mismatch")
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "invalid_state",
			Message: "OAuth state mismatch",
		})
	}

	// Clear state cookie
	clearCookie := &http.Cookie{
		Name:   "oauth_state",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	}
	c.SetCookie(clearCookie)

	// Exchange code for token
	code := c.QueryParam("code")
	if code == "" {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "missing_code",
			Message: "Authorization code missing",
		})
	}

	token, err := h.oauthConfig.Exchange(context.Background(), code)
	if err != nil {
		h.log.WithError(err).Error("Failed to exchange OAuth code")
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: "exchange_failed",
			Message: "Failed to exchange authorization code",
		})
	}

	// Fetch user info from Discord
	discordUser, err := h.fetchDiscordUser(token.AccessToken)
	if err != nil {
		h.log.WithError(err).Error("Failed to fetch Discord user")
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: "user_fetch_failed",
			Message: "Failed to fetch user information",
		})
	}

	// Fetch guild member info to get roles
	guildMember, err := h.fetchGuildMember(token.AccessToken, discordUser.ID)
	if err != nil {
		h.log.WithError(err).Error("Failed to fetch guild member")
		return c.JSON(http.StatusForbidden, dto.ErrorResponse{
			Error: "guild_access_denied",
			Message: "You must be a member of the Sol Armada guild",
		})
	}

	// Map Discord roles to application roles
	userRoles := h.mapDiscordRoles(guildMember.Roles)

	// Use Discord ID directly as user ID (no database user table)
	userID := discordUser.ID
	avatarURL := fmt.Sprintf("https://cdn.discordapp.com/avatars/%s/%s.png", discordUser.ID, discordUser.Avatar)

	// Generate JWT token
	jwtToken, err := h.tokenService.GenerateToken(
		userID,
		discordUser.ID,
		discordUser.Username,
		discordUser.Email,
		userRoles,
	)
	if err != nil {
		h.log.WithError(err).Error("Failed to generate JWT")
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "token_generation_failed",
			Message: "Failed to create session",
		})
	}

	// Create session in Redis
	ctx := context.Background()
	sessionExpiry := 7 * 24 // 7 days in hours
	_, err = h.sessionService.CreateSession(ctx, userID, jwtToken, sessionExpiry)
	if err != nil {
		h.log.WithError(err).Error("Failed to create session")
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "session_creation_failed",
			Message: "Failed to create session",
		})
	}

	// Set session cookie
	h.cookieService.SetSessionCookie(c, jwtToken, sessionExpiry*60*60) // Convert hours to seconds

	// Generate and set CSRF token
	csrfToken := auth.GenerateCSRFToken()
	h.cookieService.SetCSRFCookie(c, csrfToken, 60*60) // 1 hour

	// Return user info and CSRF token
	response := dto.AuthResponse{
		User: dto.UserDTO{
			ID:        userID,
			DiscordID: discordUser.ID,
			Username:  discordUser.Username,
			Email:     discordUser.Email,
			Avatar:    avatarURL,
			Roles:     userRoles,
		},
		CSRF: csrfToken,
	}

	return c.JSON(http.StatusOK, response)
}

// Logout clears the session
func (h *AuthHandler) Logout(c echo.Context) error {
	h.cookieService.ClearSessionCookie(c)
	h.cookieService.ClearCSRFCookie(c)
	return c.JSON(http.StatusOK, map[string]string{"message": "Logged out successfully"})
}

// Me returns the current user info
func (h *AuthHandler) Me(c echo.Context) error {
	// Extract user info from context (set by auth middleware)
	userID, _ := c.Get("user_id").(string)
	discordID, _ := c.Get("discord_id").(string)
	username, _ := c.Get("username").(string)
	email, _ := c.Get("email").(string)
	roles, _ := c.Get("roles").([]string)

	user := dto.UserDTO{
		ID:        userID,
		DiscordID: discordID,
		Username:  username,
		Email:     email,
		Roles:     roles,
	}

	return c.JSON(http.StatusOK, user)
}

// fetchDiscordUser fetches user info from Discord API
func (h *AuthHandler) fetchDiscordUser(accessToken string) (*dto.DiscordUser, error) {
	req, err := http.NewRequest("GET", DiscordAPIURL+"/users/@me", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("discord API error: %s", string(body))
	}

	var user dto.DiscordUser
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, err
	}

	return &user, nil
}

// fetchGuildMember fetches guild member info from Discord API
func (h *AuthHandler) fetchGuildMember(accessToken, userID string) (*dto.DiscordGuildMember, error) {
	url := fmt.Sprintf("%s/users/@me/guilds/%s/member", DiscordAPIURL, h.guildID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("guild member fetch error: %s", string(body))
	}

	var member dto.DiscordGuildMember
	if err := json.NewDecoder(resp.Body).Decode(&member); err != nil {
		return nil, err
	}

	return &member, nil
}

// mapDiscordRoles maps Discord role IDs to application roles
func (h *AuthHandler) mapDiscordRoles(discordRoles []string) []string {
	roles := []string{"member"} // Everyone is at least a member

	for _, roleID := range discordRoles {
		if roleID == h.adminRoleID {
			roles = append(roles, "admin")
		}
		if roleID == h.modRoleID {
			roles = append(roles, "moderator")
		}
	}

	return roles
}

// generateRandomState generates a random state string for OAuth
func generateRandomState() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}
