package models

import (
	"time"

	"github.com/google/uuid"
)

type Chat struct {
	ID             uuid.UUID `db:"chat_id"`
	Title          string    `db:"chat_title"`
	CreatedAt      time.Time `db:"chat_created_at"`
	UpdatedAt      time.Time `db:"chat_updated_at"`
	UserID         uuid.UUID `db:"chat_user_id"`
	UnloggedUserID uuid.UUID `db:"chat_unlogged_user_id"`
}
