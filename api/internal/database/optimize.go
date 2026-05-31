package database

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"
)

// OptimizePool optimizes the connection pool configuration
func OptimizePool(pool *pgxpool.Pool, logger *logrus.Logger) {
	config := pool.Config()

	logger.WithFields(logrus.Fields{
		"max_conns":          config.MaxConns,
		"min_idle_conns":     config.MinIdleConns,
		"max_conn_idle_time": config.MaxConnIdleTime,
		"max_conn_lifetime":  config.MaxConnLifetime,
		"health_check":       config.HealthCheckPeriod,
	}).Info("Connection pool configuration")

	// Log pool stats
	stats := pool.Stat()
	logger.WithFields(logrus.Fields{
		"conns_acquired": stats.AcquiredConns,
		"conns_idle":     stats.IdleConns,
		"conns_total":    stats.TotalConns,
	}).Info("Connection pool stats")
}

// ConnectionPoolConfig returns optimized pool configuration
func ConnectionPoolConfig() map[string]interface{} {
	return map[string]interface{}{
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
func LogPoolHealth(pool *pgxpool.Pool, logger *logrus.Logger) {
	stats := pool.Stat()
	logger.WithFields(logrus.Fields{
		"acquired_conns": stats.AcquiredConns,
		"idle_conns":     stats.IdleConns,
		"total_conns":    stats.TotalConns,
	}).Debug("Connection pool health")
}
