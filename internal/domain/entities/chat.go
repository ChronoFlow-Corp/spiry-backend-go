package entities

import (
	"time"

	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/domain"
	"github.com/google/uuid"
)

const maxLenTitle = 256

// Chat is entity of commands.
type Chat struct {
	ID        uuid.UUID
	Title     string
	CreatedAt time.Time
	UpdatedAt time.Time
	UserID    *uuid.UUID
}

func NewChat(userID *uuid.UUID, title string) *Chat {
	return &Chat{
		ID:        uuid.New(),
		UserID:    userID,
		Title:     title,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func (c *Chat) Validate() error {
	if len(c.Title) > maxLenTitle {
		return domain.NewValidationError(nil, "title", "title too long")
	}

	if c.CreatedAt.After(c.UpdatedAt) {
		return domain.NewValidationError(nil, "created_at", "createdAt is in the future")
	}

	return nil
}
