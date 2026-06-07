package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Logging  LoggingConfig
	Database DatabaseConfig
	Redis    RedisConfig
	Auth     AuthConfig
	Server   ServerConfig
	Discord  DiscordConfig
	Roles    RolesConfig
}

type LoggingConfig struct {
	Level string
	HUMAN bool
}

type DatabaseConfig struct {
	DSN            string
	MaxConnections int
}

type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

type AuthConfig struct {
	JWTSecret          string
	SessionExpiryHours int
}

type ServerConfig struct {
	Port        int
	Environment string
	FrontendURL string
}

type DiscordConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURI  string
	Scopes       []string
	GuildID      string
}

type RolesConfig struct {
	AdminRoleID     string
	ModeratorRoleID string
}

// load reads configuration from environment variables
func load() (Config, error) {
	cfg := Config{
		Logging: LoggingConfig{
			Level: getEnv("LOG_LEVEL", "info"),
			HUMAN: getEnvBool("LOG_HUMAN", false),
		},
		Database: DatabaseConfig{
			DSN:            getEnv("DATABASE_DSN", "postgres://localhost/website"),
			MaxConnections: getEnvInt("DATABASE_MAX_CONNECTIONS", 20),
		},
		Redis: RedisConfig{
			Addr:     getEnv("REDIS_ADDR", "localhost:6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvInt("REDIS_DB", 0),
		},
		Auth: AuthConfig{
			JWTSecret:          getEnv("JWT_SECRET", ""),
			SessionExpiryHours: getEnvInt("SESSION_EXPIRY_HOURS", 24),
		},
		Server: ServerConfig{
			Port:        getEnvInt("SERVER_PORT", 8080),
			Environment: getEnv("SERVER_ENV", "development"),
			FrontendURL: getEnv("FRONTEND_URL", "http://localhost:3000"),
		},
		Discord: DiscordConfig{
			ClientID:     getEnv("DISCORD_CLIENT_ID", ""),
			ClientSecret: getEnv("DISCORD_CLIENT_SECRET", ""),
			RedirectURI:  getEnv("DISCORD_REDIRECT_URI", ""),
			Scopes:       getEnvSlice("DISCORD_SCOPES", []string{"identify", "guilds", "guilds.members.read"}),
			GuildID:      getEnv("DISCORD_GUILD_ID", ""),
		},
		Roles: RolesConfig{
			AdminRoleID:     getEnv("ADMIN_ROLE_ID", ""),
			ModeratorRoleID: getEnv("MODERATOR_ROLE_ID", ""),
		},
	}

	// Validate required fields
	if cfg.Auth.JWTSecret == "" {
		return cfg, fmt.Errorf("JWT_SECRET environment variable is required")
	}
	if cfg.Discord.ClientID == "" {
		return cfg, fmt.Errorf("DISCORD_CLIENT_ID environment variable is required")
	}
	if cfg.Discord.ClientSecret == "" {
		return cfg, fmt.Errorf("DISCORD_CLIENT_SECRET environment variable is required")
	}
	if cfg.Discord.GuildID == "" {
		return cfg, fmt.Errorf("DISCORD_GUILD_ID environment variable is required")
	}

	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		return strings.ToLower(value) == "true" || value == "1"
	}
	return defaultValue
}

func getEnvSlice(key string, defaultValue []string) []string {
	if value := os.Getenv(key); value != "" {
		return strings.Split(value, ",")
	}
	return defaultValue
}
