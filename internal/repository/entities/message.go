package entities

import (
	"time"

	"github.com/google/uuid"
)


const (
	UserRole = "user"
	ModelRole = "model"
)

type Message struct {
	ID        uuid.UUID
	Text      string
	URL string
	MediaType string
	Role string
	ChatID    uuid.UUID
	UserID    uuid.UUID
	CreatedAt time.Time
	UpdatedAt time.Time
}