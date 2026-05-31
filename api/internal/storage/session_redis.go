package storage

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/sol-armada/website/internal/models"
)

const (
	sessionKeyPrefix = "session:"
	userSessionsKey  = "user_sessions:"
)

// RedisSessionStorage implements SessionStorage interface with Redis
type RedisSessionStorage struct {
	client *redis.Client
}

// NewRedisSessionStorage creates a new Redis session storage
func NewRedisSessionStorage(redisClient *RedisClient) *RedisSessionStorage {
	return &RedisSessionStorage{
		client: redisClient.Client,
	}
}

// Create creates a new session in Redis
func (s *RedisSessionStorage) Create(ctx context.Context, session *models.Session) error {
	session.CreatedAt = time.Now()
	
	// Serialize session to JSON
	data, err := json.Marshal(session)
	if err != nil {
		return fmt.Errorf("failed to marshal session: %w", err)
	}
	
	// Calculate TTL
	ttl := time.Until(session.ExpiresAt)
	if ttl <= 0 {
		return errors.New("session already expired")
	}
	
	// Store session with expiry
	key := sessionKeyPrefix + session.ID
	if err := s.client.Set(ctx, key, data, ttl).Err(); err != nil {
		return fmt.Errorf("failed to store session: %w", err)
	}
	
	// Add to user's session set
	userKey := userSessionsKey + session.UserID
	if err := s.client.SAdd(ctx, userKey, session.ID).Err(); err != nil {
		return fmt.Errorf("failed to add session to user set: %w", err)
	}
	
	// Set expiry on user's session set
	if err := s.client.Expire(ctx, userKey, ttl+24*time.Hour).Err(); err != nil {
		return fmt.Errorf("failed to set expiry on user session set: %w", err)
	}
	
	return nil
}

// GetByID retrieves a session by ID from Redis
func (s *RedisSessionStorage) GetByID(ctx context.Context, id string) (*models.Session, error) {
	key := sessionKeyPrefix + id
	
	data, err := s.client.Get(ctx, key).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, ErrSessionNotFound
		}
		return nil, fmt.Errorf("failed to get session: %w", err)
	}
	
	var session models.Session
	if err := json.Unmarshal(data, &session); err != nil {
		return nil, fmt.Errorf("failed to unmarshal session: %w", err)
	}
	
	// Check if expired (Redis should handle this, but double-check)
	if time.Now().After(session.ExpiresAt) {
		s.Delete(ctx, id)
		return nil, ErrSessionNotFound
	}
	
	return &session, nil
}

// GetByUserID retrieves active sessions for a user from Redis
func (s *RedisSessionStorage) GetByUserID(ctx context.Context, userID string) ([]*models.Session, error) {
	userKey := userSessionsKey + userID
	
	// Get all session IDs for this user
	sessionIDs, err := s.client.SMembers(ctx, userKey).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get user sessions: %w", err)
	}
	
	var sessions []*models.Session
	for _, sessionID := range sessionIDs {
		session, err := s.GetByID(ctx, sessionID)
		if err != nil {
			if errors.Is(err, ErrSessionNotFound) {
				// Session expired, remove from set
				s.client.SRem(ctx, userKey, sessionID)
				continue
			}
			return nil, err
		}
		sessions = append(sessions, session)
	}
	
	return sessions, nil
}

// Delete deletes a session from Redis
func (s *RedisSessionStorage) Delete(ctx context.Context, id string) error {
	key := sessionKeyPrefix + id
	
	// Get session to find user ID
	session, err := s.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, ErrSessionNotFound) {
			return nil // Already deleted
		}
		return err
	}
	
	// Remove from Redis
	if err := s.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}
	
	// Remove from user's session set
	userKey := userSessionsKey + session.UserID
	if err := s.client.SRem(ctx, userKey, id).Err(); err != nil {
		return fmt.Errorf("failed to remove session from user set: %w", err)
	}
	
	return nil
}

// DeleteByUserID deletes all sessions for a user from Redis
func (s *RedisSessionStorage) DeleteByUserID(ctx context.Context, userID string) error {
	userKey := userSessionsKey + userID
	
	// Get all session IDs
	sessionIDs, err := s.client.SMembers(ctx, userKey).Result()
	if err != nil {
		return fmt.Errorf("failed to get user sessions: %w", err)
	}
	
	// Delete each session
	for _, sessionID := range sessionIDs {
		key := sessionKeyPrefix + sessionID
		if err := s.client.Del(ctx, key).Err(); err != nil {
			return fmt.Errorf("failed to delete session %s: %w", sessionID, err)
		}
	}
	
	// Delete user's session set
	if err := s.client.Del(ctx, userKey).Err(); err != nil {
		return fmt.Errorf("failed to delete user session set: %w", err)
	}
	
	return nil
}

// DeleteExpired is a no-op for Redis (Redis handles expiration automatically)
func (s *RedisSessionStorage) DeleteExpired(ctx context.Context) (int64, error) {
	// Redis automatically expires keys, so this is not needed
	return 0, nil
}
