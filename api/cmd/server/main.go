package main

import (
	"context"
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

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/sirupsen/logrus"
	"github.com/sol-armada/sol-bot/attendance"
	solbotdb "github.com/sol-armada/sol-bot/database"
	solbotpg "github.com/sol-armada/sol-bot/database/postgresql"
	"github.com/sol-armada/sol-bot/members"
	"github.com/sol-armada/sol-bot/tokens"

	"github.com/sol-armada/website/internal/auth"
	"github.com/sol-armada/website/internal/cache"
	"github.com/sol-armada/website/internal/database"
	"github.com/sol-armada/website/internal/handlers"
	appMiddleware "github.com/sol-armada/website/internal/middleware"
	"github.com/sol-armada/website/internal/service"
	"github.com/sol-armada/website/internal/storage"
)

var (
	version = "local"
	hash    = "local"
)

func main() {
	log := logrus.New()
	log.SetFormatter(&logrus.JSONFormatter{})

	// Load configuration
	cfg, err := Load()
	if err != nil {
		log.WithError(err).Fatal("Failed to load configuration")
	}

	// Set log level
	level, err := logrus.ParseLevel(cfg.Logging.Level)
	if err != nil {
		log.WithError(err).Fatal("Invalid log level")
	}
	log.SetLevel(level)

	log.WithFields(logrus.Fields{
		"version": version,
		"hash":    hash,
		"env":     cfg.Server.Environment,
	}).Info("Starting Sol Armada Website API")

	// Initialize database connection (read-only for sol-bot member data)
	solbotCfg, err := toSolBotPostgresConfig(cfg.Database.DSN, cfg.Database.MaxConnections)
	if err != nil {
		log.WithError(err).Fatal("Invalid database DSN")
	}

	solbotClient, err := solbotpg.New(context.Background(), solbotCfg)
	if err != nil {
		log.WithError(err).Fatal("Failed to initialize sol-bot postgresql client")
	}
	defer solbotClient.Close()

	if err := members.Setup(); err != nil {
		log.WithError(err).Fatal("Failed to initialize sol-bot members backend")
	}
	if err := attendance.Setup(); err != nil {
		log.WithError(err).Fatal("Failed to initialize sol-bot attendance backend")
	}
	if err := tokens.Setup(); err != nil {
		log.WithError(err).Fatal("Failed to initialize sol-bot tokens backend")
	}

	// Initialize Redis connection for sessions
	redisConfig := storage.RedisConfig{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	}
	
	redisClient, err := storage.NewRedisClient(redisConfig, log)
	if err != nil {
		log.WithError(err).Fatal("Failed to connect to Redis")
	}
	defer redisClient.Close()

	// Initialize cache layer
	redisCache, err := cache.NewRedisCache(cfg.Redis.Addr, log)
	if err != nil {
		log.WithError(err).Warn("Failed to initialize Redis cache (continuing without caching)")
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
		if err := database.ExecuteOptimizations(solbotClient.Pool, log); err != nil {
			log.WithError(err).Warn("Failed to execute database optimizations")
		}
	}

	// Add rate limiting for API routes (10 requests per second, burst 20)
	apiRateLimiter := appMiddleware.NewRateLimiter(10, 20)

	// Health check endpoint
	e.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]interface{}{
			"status":  "ok",
			"version": version,
			"hash":    hash,
		})
	})

	// Version endpoint
	e.GET("/version", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]interface{}{
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

	adminAPI := api.Group("/admin")
	adminAPI.GET("/overview", adminHandler.GetOverview)
	adminAPI.GET("/attendance", adminHandler.GetAttendance)
	adminAPI.GET("/token-ledger", adminHandler.GetTokenLedger)
	adminAPI.GET("/members", adminHandler.GetMembers)

	// Start server in a goroutine
	go func() {
		log.WithField("port", cfg.Server.Port).Info("Server listening")
		if err := e.Start(":" + fmt.Sprintf("%d", cfg.Server.Port)); err != nil && err != http.ErrServerClosed {
			log.WithError(err).Error("Server error")
		}
	}()

	// Graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Info("Shutting down server")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := e.Shutdown(ctx); err != nil {
		log.WithError(err).Fatal("Server shutdown error")
	}

	log.Info("Server stopped")
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
