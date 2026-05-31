package database

import (
	"context"
	"time"

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

// QueryOptimizations provides index recommendations
var QueryOptimizations = map[string]string{
	"attendance_by_date": "CREATE INDEX IF NOT EXISTS idx_attendance_date_created ON attendance(date_created DESC)",
	"members_by_rank":    "CREATE INDEX IF NOT EXISTS idx_members_rank ON members(rank)",
	"tokens_by_member":   "CREATE INDEX IF NOT EXISTS idx_tokens_member_id ON tokens(member_id)",
	"tokens_by_created":  "CREATE INDEX IF NOT EXISTS idx_tokens_created_at ON tokens(created_at DESC)",
}

// GetIndexRecommendations returns SQL for recommended indexes
func GetIndexRecommendations() []string {
	recommendations := make([]string, 0, len(QueryOptimizations))
	for _, sql := range QueryOptimizations {
		recommendations = append(recommendations, sql)
	}
	return recommendations
}

// ExecuteOptimizations applies recommended indexes
func ExecuteOptimizations(pool *pgxpool.Pool, logger *logrus.Logger) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	for name, sql := range QueryOptimizations {
		if _, err := pool.Exec(ctx, sql); err != nil {
			logger.WithError(err).Warnf("Failed to create index: %s", name)
			continue
		}
		logger.Infof("Created index: %s", name)
	}

	return nil
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
