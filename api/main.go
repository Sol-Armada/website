package main

import (
	"context"
	"embed"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"log/slog"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/sol-armada/sol-bot/attendance"
	solbotdb "github.com/sol-armada/sol-bot/database"
	"github.com/sol-armada/sol-bot/database/dbnotify"
	solbotpg "github.com/sol-armada/sol-bot/database/postgresql"
	"github.com/sol-armada/sol-bot/members"
	"github.com/sol-armada/sol-bot/tokens"

	"github.com/sol-armada/website/internal/auth"
	"github.com/sol-armada/website/internal/cache"
	"github.com/sol-armada/website/internal/database"
	"github.com/sol-armada/website/internal/handlers"
	appMiddleware "github.com/sol-armada/website/internal/middleware"
	"github.com/sol-armada/website/internal/realtime"
	"github.com/sol-armada/website/internal/service"
	"github.com/sol-armada/website/internal/storage"
)

//go:embed dist
var frontendFS embed.FS

var (
	version = "local"
	hash    = "local"
)

func main() {
	// Load configuration first
	cfg, err := load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Parse log level
	var logLevel slog.Level
	if err := logLevel.UnmarshalText([]byte(cfg.Logging.Level)); err != nil {
		logLevel = slog.LevelInfo
	}

	var handler slog.Handler
	handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: logLevel,
	})
	if cfg.Logging.HUMAN {
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: logLevel,
		})
	}

	// Create logger with JSON handler and parsed level
	log := slog.New(handler)

	log.Info("Starting Sol Armada Website API",
		"version", version,
		"hash", hash,
		"env", cfg.Server.Environment,
	)

	// Initialize database connection
	solbotCfg, err := toSolBotPostgresConfig(cfg.Database.DSN, cfg.Database.MaxConnections)
	if err != nil {
		log.Error("Invalid database DSN", "error", err)
		os.Exit(1)
	}

	solbotClient, err := solbotpg.New(context.Background(), solbotCfg)
	if err != nil {
		log.Error("Failed to initialize sol-bot postgresql client", "error", err)
		os.Exit(1)
	}
	defer solbotClient.Close()

	if err := members.Setup(); err != nil {
		log.Error("Failed to initialize sol-bot members backend", "error", err)
		os.Exit(1)
	}
	if err := attendance.Setup(); err != nil {
		log.Error("Failed to initialize sol-bot attendance backend", "error", err)
		os.Exit(1)
	}
	if err := tokens.Setup(); err != nil {
		log.Error("Failed to initialize sol-bot tokens backend", "error", err)
		os.Exit(1)
	}

	// Initialize Redis connection for sessions
	redisConfig := storage.RedisConfig{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	}

	redisClient, err := storage.NewRedisClient(redisConfig, log)
	if err != nil {
		log.Error("Failed to connect to Redis", "error", err)
		os.Exit(1)
	}
	defer redisClient.Close()

	// Initialize cache layer
	redisCache, err := cache.NewRedisCache(cfg.Redis.Addr, log)
	if err != nil {
		log.Warn("Failed to initialize Redis cache (continuing without caching)", "error", err)
		redisCache = nil
	}
	if redisCache != nil {
		defer redisCache.Close()
	}

	// Initialize storage layer
	sessionStorage := storage.NewRedisSessionStorage(redisClient)

	// Initialize services
	sessionService := service.NewSessionService(sessionStorage, log)
	memberService := service.NewMemberService(log)
	adminService := service.NewAdminService(log)

	// Wrap admin service with caching if Redis is available
	var adminServiceInterface handlers.AdminServiceInterface = adminService

	if redisCache != nil {
		cachedAdminService := service.NewCachedAdminService(adminService, redisCache, log)
		adminServiceInterface = cachedAdminService
		log.Info("Admin service caching enabled")
	} else {
		log.Warn("Admin service caching disabled")
	}

	// Initialize auth services
	tokenService := auth.NewTokenService(
		cfg.Auth.JWTSecret,
		"sol-armada-api",
		cfg.Auth.SessionExpiryHours,
	)
	cookieService := auth.NewCookieService(cfg.Server.Environment, "")

	// Initialize middleware
	authMiddleware := appMiddleware.NewAuthMiddleware(tokenService, cookieService)

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(
		cfg.Discord.ClientID,
		cfg.Discord.ClientSecret,
		cfg.Discord.RedirectURI,
		cfg.Server.FrontendURL,
		cfg.Discord.Scopes,
		tokenService,
		cookieService,
		sessionService,
		cfg.Discord.GuildID,
		cfg.Roles.AdminRoleID,
		cfg.Roles.ModeratorRoleID,
		log,
	)
	memberHandler := handlers.NewMemberHandler(memberService, log)
	adminHandler := handlers.NewAdminHandler(adminServiceInterface, log)
	wsHub := realtime.NewHub(log)
	go wsHub.RunHealthHeartbeat(20 * time.Second)
	wsHandler := handlers.NewWebSocketHandler(wsHub, log)

	notifyListener, err := dbnotify.NewListener(dbnotify.ListenerConfig{
		DSN: cfg.Database.DSN,
		Channels: []string{
			dbnotify.ChannelMembers,
			dbnotify.ChannelAttendance,
			dbnotify.ChannelTokens,
		},
		OnError: func(listenerErr error) {
			log.Warn("db notify listener warning", "error", listenerErr)
		},
	})
	if err != nil {
		log.Error("Failed to initialize db notify listener", "error", err)
		os.Exit(1)
	}

	notifyCtx, cancelNotify := context.WithCancel(context.Background())
	go func() {
		if runErr := notifyListener.Run(notifyCtx, func(_ context.Context, event dbnotify.Event) error {
			topic, ok := realtime.TopicForNotifyChannel(event.Channel)
			if !ok {
				return nil
			}

			payload := map[string]any{
				"channel":         event.Channel,
				"operation":       event.Operation,
				"schema":          event.Schema,
				"table":           event.Table,
				"primary_key":     event.PrimaryKey,
				"occurred_at":     event.OccurredAt,
				"changed_columns": event.ChangedColumns,
			}

			if topic == realtime.TopicAdminMembers {
				memberID := extractPrimaryKeyID(event.PrimaryKey)
				if memberID != "" {
					payload["member_id"] = memberID
				}

				operation := strings.ToLower(event.Operation)
				if operation != "delete" && memberID != "" {
					memberSummary, memberErr := adminService.GetMemberSummaryByID(notifyCtx, memberID)
					if memberErr != nil {
						log.Debug("failed to enrich member ws payload", "member_id", memberID, "error", memberErr)
					} else if memberSummary != nil {
						payload["member"] = memberSummary
					}
				}
			}

			wsHub.Publish(topic, payload)

			return nil
		}); runErr != nil {
			log.Warn("db notify listener stopped", "error", runErr)
		}
	}()

	// Setup Echo router
	e := echo.New()
	e.HideBanner = true

	// Add global middleware
	e.Use(appMiddleware.LoggingMiddleware(log))
	e.Use(appMiddleware.ErrorLoggerMiddleware(log))
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	// Database optimization
	if solbotClient != nil && solbotClient.Pool != nil {
		database.OptimizePool(solbotClient.Pool, log)
	}

	// Add rate limiting for API routes (10 requests per second, burst 20)
	apiRateLimiter := appMiddleware.NewRateLimiter(10, 20)

	// Health check endpoint
	e.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]any{
			"status":  "ok",
			"version": version,
			"hash":    hash,
		})
	})

	// Version endpoint
	e.GET("/version", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]any{
			"version": version,
			"hash":    hash,
		})
	})

	// Auth routes
	authGroup := e.Group("/auth")
	authGroup.GET("/login", authHandler.Login)
	authGroup.GET("/callback", authHandler.Callback)
	authGroup.POST("/logout", authHandler.Logout, authMiddleware.RequireAuth)
	authGroup.GET("/me", authHandler.Me, authMiddleware.RequireAuth)

	// API routes (protected)
	api := e.Group("/api")
	api.Use(authMiddleware.RequireAuth)
	api.Use(apiRateLimiter.Middleware())

	memberAPI := api.Group("/member")
	memberAPI.GET("/dashboard", memberHandler.GetDashboard)
	memberAPI.GET("/profile", memberHandler.GetProfile)
	memberAPI.GET("/token-ledger", memberHandler.GetTokenLedger)

	adminAPI := api.Group("/admin")
	adminAPI.GET("/overview", adminHandler.GetOverview)
	adminAPI.GET("/attendance", adminHandler.GetAttendance)
	adminAPI.GET("/token-ledger", adminHandler.GetTokenLedger)
	adminAPI.GET("/token-ledger/analytics", adminHandler.GetTokenLedgerAnalytics)
	adminAPI.GET("/members", adminHandler.GetMembers)
	api.GET("/ws", wsHandler.Handle)

	// Static file serving (SPA with embedded frontend)
	staticHandler := handlers.NewStaticHandler(frontendFS, log)
	e.GET("/*", staticHandler.Handle)

	// Start server in a goroutine
	go func() {
		log.Info("Server listening", "port", cfg.Server.Port)
		if err := e.Start(":" + fmt.Sprintf("%d", cfg.Server.Port)); err != nil && err != http.ErrServerClosed {
			log.Error("Server error", "error", err)
		}
	}()

	// Graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Info("Shutting down server")
	cancelNotify()
	wsHub.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := e.Shutdown(ctx); err != nil {
		log.Error("Server shutdown error", "error", err)
		os.Exit(1)
	}

	log.Info("Server stopped")
}

func extractPrimaryKeyID(primaryKey map[string]any) string {
	if len(primaryKey) == 0 {
		return ""
	}

	if idValue, ok := primaryKey["id"]; ok && idValue != nil {
		return fmt.Sprint(idValue)
	}

	for _, value := range primaryKey {
		if value != nil {
			return fmt.Sprint(value)
		}
	}

	return ""
}

func toSolBotPostgresConfig(dsn string, maxConns int) (solbotdb.PostgresConfig, error) {
	u, err := url.Parse(dsn)
	if err != nil {
		return solbotdb.PostgresConfig{}, fmt.Errorf("parse dsn: %w", err)
	}

	host := u.Hostname()
	if host == "" {
		return solbotdb.PostgresConfig{}, fmt.Errorf("dsn host is empty")
	}

	port := 5432
	if rawPort := u.Port(); rawPort != "" {
		parsedPort, err := strconv.Atoi(rawPort)
		if err != nil {
			return solbotdb.PostgresConfig{}, fmt.Errorf("invalid dsn port: %w", err)
		}
		port = parsedPort
	} else if tcpAddr, err := net.LookupPort("tcp", "postgres"); err == nil {
		port = tcpAddr
	}

	databaseName := strings.TrimPrefix(u.Path, "/")
	if databaseName == "" {
		return solbotdb.PostgresConfig{}, fmt.Errorf("dsn database name is empty")
	}

	username := ""
	password := ""
	if u.User != nil {
		username = u.User.Username()
		password, _ = u.User.Password()
	}

	if username == "" {
		return solbotdb.PostgresConfig{}, fmt.Errorf("dsn username is empty")
	}

	sslMode := u.Query().Get("sslmode")
	if sslMode == "" {
		sslMode = "disable"
	}

	return solbotdb.PostgresConfig{
		Host:     host,
		Port:     port,
		Username: username,
		Password: password,
		Database: databaseName,
		SSLMode:  sslMode,
		MaxConns: int32(maxConns),
		MinConns: 1,
	}, nil
}
