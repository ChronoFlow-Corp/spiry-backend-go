package entities

import (
	"time"

	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/domain"
	"github.com/google/uuid"
)

const (
	maxLenModelName = 128
)

type Model struct {
	ID        uuid.UUID
	Name      string
	MinLevel  uint
	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewModel(name string, minLevel uint) *Model {
	return &Model{
		ID:        uuid.New(),
		Name:      name,
		MinLevel:  minLevel,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now().Add(time.Millisecond),
	}
}

func (m *Model) Validate() error {
	if len(m.Name) > maxLenModelName {
		return domain.NewValidationError(nil, "name", "name is too long")
	}

	if m.CreatedAt.After(m.UpdatedAt) {
		return domain.NewValidationError(nil, "created_at", "createdAt is in the future")
	}

	return nil
}
