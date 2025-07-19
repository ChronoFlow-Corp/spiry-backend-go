package postgres

import (
	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/repository"
	"github.com/google/uuid"
	"time"
)

type user struct {
	ID        uuid.UUID `db:"id"`
	Email     string    `db:"email"`
	AccessTokenGoogle string    `db:"access_token_google"`
	RefreshTokenGoogle string    `db:"refresh_token_google"`
	RefreshToken      string    `db:"refresh_token"`
	CreatedAt         time.Time `db:"created_at"`
	UpdatedAt         time.Time `db:"updated_at"`
}

func toUser(u user) repository.User {
	return repository.NewUser(u.ID, u.Email, u.AccessTokenGoogle, u.RefreshTokenGoogle, u.RefreshToken)
}