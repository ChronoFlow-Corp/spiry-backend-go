package entities

import (
	"time"

	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/domain"
	"github.com/google/uuid"
)

type Session struct {
	ID        uuid.UUID
	Token     string
	ExpiresAt time.Time
	LastLogin time.Time
	UserID    *uuid.UUID
	Device    string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewSession(lastLogin, expiresAt time.Time, userID *uuid.UUID, token, device string) *Session {
	return &Session{
		ID:        uuid.New(),
		LastLogin: lastLogin,
		Token:     token,
		ExpiresAt: expiresAt,
		UserID:    userID,
		Device:    device,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().Add(time.Millisecond).UTC(),
	}
}

func (s *Session) Validate() error {
	if s.Token == "" {
		return domain.NewValidationError(nil, "token", "token is required")
	}

	if s.ExpiresAt.Before(time.Now()) {
		return domain.NewValidationError(nil, "token", "token is expired")
	}

	if s.Device == "" {
		return domain.NewValidationError(nil, "device", "device is required")
	}

	if s.CreatedAt.Unix() > s.UpdatedAt.Unix() {
		return domain.NewValidationError(nil, "created_at", "createdAt is in the future")
	}

	return nil
}
