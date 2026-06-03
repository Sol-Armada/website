package dto

// AuthResponse is returned after successful authentication
type AuthResponse struct {
	User UserDTO `json:"user"`
	CSRF string  `json:"csrf_token"`
}

// UserDTO represents user data for API responses
type UserDTO struct {
	ID        string   `json:"id"`
	DiscordID string   `json:"discord_id"`
	Username  string   `json:"username"`
	Email     string   `json:"email"`
	Avatar    string   `json:"avatar,omitempty"`
	Roles     []string `json:"roles"`
}

// DiscordUser represents user data from Discord API
type DiscordUser struct {
	ID            string `json:"id"`
	Username      string `json:"username"`
	Discriminator string `json:"discriminator"`
	Avatar        string `json:"avatar"`
	Email         string `json:"email"`
	Verified      bool   `json:"verified"`
}

// DiscordGuild represents guild data from Discord API
type DiscordGuild struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Icon        string `json:"icon"`
	Owner       bool   `json:"owner"`
	Permissions string `json:"permissions"`
}

// DiscordGuildMember represents guild member data from Discord API
type DiscordGuildMember struct {
	User         *DiscordUser `json:"user"`
	Nick         string       `json:"nick"`
	Avatar       string       `json:"avatar"`
	Roles        []string     `json:"roles"`
	JoinedAt     string       `json:"joined_at"`
	PremiumSince string       `json:"premium_since"`
}

// ErrorResponse represents an API error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}
