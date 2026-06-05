package database

import (
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
)

// OptimizePool optimizes the connection pool configuration
func OptimizePool(pool *pgxpool.Pool, logger *slog.Logger) {
	config := pool.Config()

	logger.Info("Connection pool configuration",
		"max_conns", config.MaxConns,
		"min_idle_conns", config.MinIdleConns,
		"max_conn_idle_time", config.MaxConnIdleTime,
		"max_conn_lifetime", config.MaxConnLifetime,
		"health_check", config.HealthCheckPeriod,
	)

	// Log pool stats
	stats := pool.Stat()
	logger.Info("Connection pool stats",
		"conns_acquired", stats.AcquiredConns(),
		"conns_idle", stats.IdleConns(),
		"conns_total", stats.TotalConns(),
	)
}

// ConnectionPoolConfig returns optimized pool configuration
func ConnectionPoolConfig() map[string]any {
	return map[string]any{
		"MaxConns":          25,
		"MinIdleConns":      5,
		"MaxConnIdleTime":   "5m",
		"MaxConnLifetime":   "1h",
		"HealthCheckPeriod": "1m",
		"ReadTimeout":       "30s",
		"WriteTimeout":      "30s",
	}
}

// LogPoolHealth periodically logs pool health
func LogPoolHealth(pool *pgxpool.Pool, logger *slog.Logger) {
	stats := pool.Stat()
	logger.Debug("Connection pool health",
		"acquired_conns", stats.AcquiredConns,
		"idle_conns", stats.IdleConns,
		"total_conns", stats.TotalConns,
	)
}
