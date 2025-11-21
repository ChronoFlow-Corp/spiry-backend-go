package models

import (
	"time"

	"github.com/google/uuid"
)

type Session struct {
	ID             uuid.UUID  `db:"id"`
	Token          string     `db:"token"`
	ExpiresAt      time.Time  `db:"expires_at"`
	LastLogin      time.Time  `db:"last_login"`
	Device         string     `db:"device"`
	UserID         *uuid.UUID `db:"user_id"`
	UnloggedUserID *uuid.UUID `db:"unlogged_user_id"`
	CreatedAt      time.Time  `db:"created_at"`
	UpdatedAt      time.Time  `db:"updated_at"`
}
