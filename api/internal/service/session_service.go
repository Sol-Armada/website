package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	
	"github.com/sol-armada/website/internal/models"
	"github.com/sol-armada/website/internal/storage"
)

// SessionService handles session-related business logic
type SessionService struct {
	sessionStorage storage.SessionStorage
	logger         *logrus.Logger
}

// NewSessionService creates a new session service
func NewSessionService(
	sessionStorage storage.SessionStorage,
	logger *logrus.Logger,
) *SessionService {
	return &SessionService{
		sessionStorage: sessionStorage,
		logger:         logger,
	}
}

// CreateSession creates a new session for a user
func (s *SessionService) CreateSession(ctx context.Context, userID, token string, expiryHours int) (*models.Session, error) {
	session := &models.Session{
		ID:        uuid.New().String(),
		UserID:    userID,
		Token:     token,
		ExpiresAt: time.Now().Add(time.Duration(expiryHours) * time.Hour),
	}
	
	if err := s.sessionStorage.Create(ctx, session); err != nil {
		s.logger.WithError(err).WithField("user_id", userID).Error("Failed to create session")
		return nil, fmt.Errorf("failed to create session: %w", err)
	}
	
	s.logger.WithFields(logrus.Fields{
		"user_id":    userID,
		"session_id": session.ID,
	}).Info("Session created")
	
	return session, nil
}

// GetSession retrieves a session by ID
func (s *SessionService) GetSession(ctx context.Context, sessionID string) (*models.Session, error) {
	return s.sessionStorage.GetByID(ctx, sessionID)
}

// DeleteSession deletes a session
func (s *SessionService) DeleteSession(ctx context.Context, sessionID string) error {
	if err := s.sessionStorage.Delete(ctx, sessionID); err != nil {
		s.logger.WithError(err).WithField("session_id", sessionID).Error("Failed to delete session")
		return fmt.Errorf("failed to delete session: %w", err)
	}
	
	s.logger.WithField("session_id", sessionID).Info("Session deleted")
	return nil
}

// DeleteUserSessions deletes all sessions for a user (logout all devices)
func (s *SessionService) DeleteUserSessions(ctx context.Context, userID string) error {
	if err := s.sessionStorage.DeleteByUserID(ctx, userID); err != nil {
		s.logger.WithError(err).WithField("user_id", userID).Error("Failed to delete user sessions")
		return fmt.Errorf("failed to delete user sessions: %w", err)
	}
	
	s.logger.WithField("user_id", userID).Info("User sessions deleted")
	return nil
}

// GetUserSessions retrieves all active sessions for a user
func (s *SessionService) GetUserSessions(ctx context.Context, userID string) ([]*models.Session, error) {
	sessions, err := s.sessionStorage.GetByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user sessions: %w", err)
	}
	return sessions, nil
}

