package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/sirupsen/logrus"

	"github.com/sol-armada/website/internal/auth"
	"github.com/sol-armada/website/internal/handlers"
	appMiddleware "github.com/sol-armada/website/internal/middleware"
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

	// TODO: Initialize database connection

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
		cfg.Discord.GuildID,
		cfg.Roles.AdminRoleID,
		cfg.Roles.ModeratorRoleID,
		log,
	)

	// Setup Echo router
	e := echo.New()
	e.HideBanner = true

	// Add global middleware
	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: `{"time":"${time_rfc3339_nano}","method":"${method}","path":"${path}","status":${status},"latency_ms":${latency_ms},"error":"${error}"}\n`,
	}))
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

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
	
	// TODO: Add API route handlers here

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
