package storage

import (
	"context"
	"errors"

	"github.com/sol-armada/website/internal/models"
)

var (
	ErrSessionNotFound = errors.New("session not found")
)

// SessionStorage defines the interface for session data persistence
type SessionStorage interface {
	Create(ctx context.Context, session *models.Session) error
	GetByID(ctx context.Context, id string) (*models.Session, error)
	GetByUserID(ctx context.Context, userID string) ([]*models.Session, error)
	Delete(ctx context.Context, id string) error
	DeleteByUserID(ctx context.Context, userID string) error
	DeleteExpired(ctx context.Context) (int64, error)
}
