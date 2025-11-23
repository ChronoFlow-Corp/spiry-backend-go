package entities

import (
	"time"

	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/domain"
	"github.com/google/uuid"
)

const (
	maxLenPassword = 128
	minLenPassword = 8
	maxLenLogin    = 64
)

// Admin is who has permission to change tools.
type Admin struct {
	ID        uuid.UUID
	Login     string
	Password  string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewAdmin(login string, password string) *Admin {
	return &Admin{
		ID:        uuid.New(),
		Login:     login,
		Password:  password,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func (a *Admin) Validate() error {
	if a.Login == "" {
		return domain.NewValidationError(nil, "login", "login is required")
	}

	if len(a.Login) > maxLenLogin {
		return domain.NewValidationError(nil, "login", "login is too long")
	}

	if len(a.Password) < minLenPassword {
		return domain.NewValidationError(nil, "password", "password is too short")
	}

	if len(a.Password) > maxLenPassword {
		return domain.NewValidationError(nil, "password", "password is too long")
	}

	if a.CreatedAt.After(a.UpdatedAt) {
		return domain.NewValidationError(nil, "created_at", "createdAt is in the future")
	}

	return nil
}
