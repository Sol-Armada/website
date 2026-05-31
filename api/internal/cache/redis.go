package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

// RedisCache wraps Redis operations
type RedisCache struct {
	client *redis.Client
	logger *logrus.Logger
}

// NewRedisCache creates a new Redis cache
func NewRedisCache(addr string, logger *logrus.Logger) (*RedisCache, error) {
	client := redis.NewClient(&redis.Options{
		Addr:            addr,
		MaxRetries:      3,
		PoolSize:        10,
		MinIdleConns:    5,
		PoolTimeout:     4 * time.Second,
		ReadTimeout:     3 * time.Second,
		WriteTimeout:    3 * time.Second,
		ConnMaxIdleTime: 5 * time.Minute,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("redis connection failed: %w", err)
	}

	logger.Info("Redis cache connected successfully")

	return &RedisCache{
		client: client,
		logger: logger,
	}, nil
}

// Get retrieves a value from cache
func (rc *RedisCache) Get(ctx context.Context, key string) (string, error) {
	return rc.client.Get(ctx, key).Result()
}

// Set stores a value in cache with TTL
func (rc *RedisCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("json marshal failed: %w", err)
	}

	return rc.client.Set(ctx, key, data, ttl).Err()
}

// GetJSON retrieves and unmarshals a JSON value from cache
func (rc *RedisCache) GetJSON(ctx context.Context, key string, result interface{}) error {
	val, err := rc.client.Get(ctx, key).Result()
	if err != nil {
		return err
	}

	return json.Unmarshal([]byte(val), result)
}

// Del deletes a key from cache
func (rc *RedisCache) Del(ctx context.Context, keys ...string) error {
	return rc.client.Del(ctx, keys...).Err()
}

// Exists checks if a key exists
func (rc *RedisCache) Exists(ctx context.Context, key string) (bool, error) {
	result, err := rc.client.Exists(ctx, key).Result()
	return result > 0, err
}

// Expire sets expiration on a key
func (rc *RedisCache) Expire(ctx context.Context, key string, expiration time.Duration) error {
	return rc.client.Expire(ctx, key, expiration).Err()
}

// FlushAll clears all cache
func (rc *RedisCache) FlushAll(ctx context.Context) error {
	return rc.client.FlushAll(ctx).Err()
}

// Close closes the Redis connection
func (rc *RedisCache) Close() error {
	return rc.client.Close()
}

// CacheKey builds standardized cache keys
type CacheKey struct {
	Prefix string
	ID     string
}

// String returns the cache key string
func (ck CacheKey) String() string {
	if ck.ID == "" {
		return ck.Prefix
	}
	return fmt.Sprintf("%s:%s", ck.Prefix, ck.ID)
}

// Common cache key prefixes
const (
	KeyMemberStats    = "member:stats"
	KeyAttendanceList = "attendance:list"
	KeyTokenBalance   = "token:balance"
	KeyAdminOverview  = "admin:overview"
)
