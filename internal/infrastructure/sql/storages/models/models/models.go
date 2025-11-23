package models

import (
	"time"

	"github.com/google/uuid"
)

type Model struct {
	ID        uuid.UUID `db:"id"`
	Name      string    `db:"name"`
	MinLevel  uint      `db:"min_level"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}
