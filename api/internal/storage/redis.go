package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

// RedisClient wraps the Redis client
type RedisClient struct {
	Client *redis.Client
	logger *logrus.Logger
}

// RedisConfig holds Redis configuration
type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

// NewRedisClient creates a new Redis client
func NewRedisClient(cfg RedisConfig, logger *logrus.Logger) (*RedisClient, error) {
	client := redis.NewClient(&redis.Options{
		Addr:         cfg.Addr,
		Password:     cfg.Password,
		DB:           cfg.DB,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		PoolSize:     10,
		MinIdleConns: 2,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	logger.WithField("addr", cfg.Addr).Info("Redis connection established")

	return &RedisClient{
		Client: client,
		logger: logger,
	}, nil
}

// Close closes the Redis connection
func (r *RedisClient) Close() error {
	r.logger.Info("Closing Redis connection")
	return r.Client.Close()
}

// Ping checks if Redis is reachable
func (r *RedisClient) Ping(ctx context.Context) error {
	return r.Client.Ping(ctx).Err()
}
