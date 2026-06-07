package handlers

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"log/slog"

	"github.com/labstack/echo/v4"
	"golang.org/x/oauth2"

	"github.com/sol-armada/website/internal/auth"
	"github.com/sol-armada/website/internal/dto"
	"github.com/sol-armada/website/internal/service"
)

const (
	DiscordAPIURL   = "https://discord.com/api/v10"
	DiscordAuthURL  = "https://discord.com/oauth2/authorize"
	DiscordTokenURL = "https://discord.com/api/oauth2/token"
)

type discordAPIError struct {
	operation  string
	statusCode int
	body       string
}

func (e *discordAPIError) Error() string {
	if e == nil {
		return "discord API error"
	}
	return fmt.Sprintf("discord %s failed with status %d: %s", e.operation, e.statusCode, e.body)
}

// AuthHandler handles authentication endpoints
type AuthHandler struct {
	oauthConfig    *oauth2.Config
	frontendURL    string
	tokenService   *auth.TokenService
	cookieService  *auth.CookieService
	sessionService *service.SessionService
	guildID        string // Required guild ID for access
	adminRoleID    string
	modRoleID      string
	log            *slog.Logger
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(
	clientID, clientSecret, redirectURI, frontendURL string,
	scopes []string,
	tokenService *auth.TokenService,
	cookieService *auth.CookieService,
	sessionService *service.SessionService,
	guildID, adminRoleID, modRoleID string,
	log *slog.Logger,
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
		frontendURL:    strings.TrimRight(frontendURL, "/"),
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
	ctx := c.Request().Context()

	// Handle OAuth provider error responses (e.g. user pressed "Cancel" on Discord consent page).
	if oauthError := c.QueryParam("error"); oauthError != "" {
		errorMessage := c.QueryParam("error_description")
		if oauthError == "access_denied" {
			errorMessage = "Authentication was cancelled in Discord"
		} else if errorMessage == "" {
			errorMessage = "Authentication failed"
		}
		h.log.Warn("OAuth callback returned error", "error", oauthError, "description", errorMessage)
		return c.Redirect(
			http.StatusFound,
			fmt.Sprintf("%s/auth/callback?error=%s&message=%s", h.frontendURL, url.QueryEscape(oauthError), url.QueryEscape(errorMessage)),
		)
	}

	// Verify state parameter
	stateCookie, err := c.Cookie("oauth_state")
	if err != nil {
		h.log.Error("Missing OAuth state cookie", "error", err)
		return c.Redirect(http.StatusFound, fmt.Sprintf("%s/auth/callback?error=invalid_state&message=%s", h.frontendURL, url.QueryEscape("OAuth state mismatch")))
	}

	state := c.QueryParam("state")
	if state == "" || state != stateCookie.Value {
		h.log.Error("OAuth state mismatch")
		return c.Redirect(http.StatusFound, fmt.Sprintf("%s/auth/callback?error=invalid_state&message=%s", h.frontendURL, url.QueryEscape("OAuth state mismatch")))
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
		return c.Redirect(http.StatusFound, fmt.Sprintf("%s/auth/callback?error=missing_code&message=%s", h.frontendURL, url.QueryEscape("Authorization code missing")))
	}

	token, err := h.oauthConfig.Exchange(ctx, code)
	if err != nil {
		h.log.Error("Failed to exchange OAuth code", "error", err)
		return c.Redirect(http.StatusFound, fmt.Sprintf("%s/auth/callback?error=exchange_failed&message=%s", h.frontendURL, url.QueryEscape("Failed to exchange authorization code")))
	}

	// Fetch user info from Discord
	discordUser, err := h.fetchDiscordUser(token.AccessToken)
	if err != nil {
		h.log.Error("Failed to fetch Discord user", "error", err)
		return c.Redirect(http.StatusFound, fmt.Sprintf("%s/auth/callback?error=user_fetch_failed&message=%s", h.frontendURL, url.QueryEscape("Failed to fetch user information")))
	}

	// Fetch guild member info to get roles
	guildMember, err := h.fetchGuildMember(token.AccessToken)
	if err != nil {
		errorCode, errorMessage := classifyGuildAccessError(err)
		h.log.Warn("Failed to fetch guild member", "error", err, "code", errorCode)
		return c.Redirect(http.StatusFound, fmt.Sprintf("%s/auth/callback?error=%s&message=%s", h.frontendURL, url.QueryEscape(errorCode), url.QueryEscape(errorMessage)))
	}

	// Map Discord roles to application roles
	userRoles := h.mapDiscordRoles(guildMember.Roles)

	// Use Discord ID directly as user ID (no database user table)
	userID := discordUser.ID

	// Generate JWT token with avatar URL
	avatarURL := fmt.Sprintf("https://cdn.discordapp.com/avatars/%s/%s.png", discordUser.ID, discordUser.Avatar)
	jwtToken, err := h.tokenService.GenerateToken(
		userID,
		discordUser.ID,
		discordUser.Username,
		discordUser.Email,
		avatarURL,
		userRoles,
	)
	if err != nil {
		h.log.Error("Failed to generate JWT", "error", err)
		return c.Redirect(http.StatusFound, fmt.Sprintf("%s/auth/callback?error=token_generation_failed&message=%s", h.frontendURL, url.QueryEscape("Failed to create session")))
	}

	// Create session in Redis
	sessionExpiry := 7 * 24 // 7 days in hours
	_, err = h.sessionService.CreateSession(ctx, userID, jwtToken, sessionExpiry)
	if err != nil {
		h.log.Error("Failed to create session", "error", err)
		return c.Redirect(http.StatusFound, fmt.Sprintf("%s/auth/callback?error=session_creation_failed&message=%s", h.frontendURL, url.QueryEscape("Failed to create session")))
	}

	// Set session cookie
	h.cookieService.SetSessionCookie(c, jwtToken, sessionExpiry*60*60) // Convert hours to seconds

	// Generate and set CSRF token
	csrfToken := auth.GenerateCSRFToken()
	h.cookieService.SetCSRFCookie(c, csrfToken, 60*60) // 1 hour

	// OAuth callback is browser-driven; redirect to frontend dashboard after cookies are set.
	return c.Redirect(http.StatusFound, h.frontendURL+"/dashboard")
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
	avatar, _ := c.Get("avatar").(string)
	roles, _ := c.Get("roles").([]string)

	user := dto.UserDTO{
		ID:        userID,
		DiscordID: discordID,
		Username:  username,
		Email:     email,
		Avatar:    avatar,
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
func (h *AuthHandler) fetchGuildMember(accessToken string) (*dto.DiscordGuildMember, error) {
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
		return nil, &discordAPIError{
			operation:  "guild member lookup",
			statusCode: resp.StatusCode,
			body:       string(body),
		}
	}

	var member dto.DiscordGuildMember
	if err := json.NewDecoder(resp.Body).Decode(&member); err != nil {
		return nil, err
	}

	return &member, nil
}

func classifyGuildAccessError(err error) (string, string) {
	var discordErr *discordAPIError
	if errors.As(err, &discordErr) {
		switch discordErr.statusCode {
		case http.StatusTooManyRequests:
			return "discord_rate_limited", "Discord is rate limiting login attempts right now. Please wait a minute and try again."
		case http.StatusNotFound:
			return "guild_access_denied", "You must be a member of the Sol Armada guild."
		case http.StatusForbidden:
			return "guild_scope_missing", "Discord denied guild membership lookup. Ask an admin to verify OAuth scopes include guilds.members.read."
		case http.StatusUnauthorized:
			return "discord_token_invalid", "Discord authentication expired. Please try signing in again."
		default:
			return "guild_lookup_failed", "Unable to verify guild membership with Discord right now. Please try again shortly."
		}
	}

	return "guild_lookup_failed", "Unable to verify guild membership with Discord right now. Please try again shortly."
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
	_, _ = rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}
