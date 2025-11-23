package models

import (
	"time"

	"github.com/google/uuid"
)

type CommandMedia struct {
	ID        uuid.UUID  `db:"id"         json:"id,omitempty"`
	Name      string     `db:"name"       json:"name"`
	Type      string     `db:"type"       json:"type"`
	URL       string     `db:"url"        json:"url"`
	Size      int        `db:"size"       json:"size"`
	CommandID *uuid.UUID `db:"command_id" json:"command_id,omitempty"`
	UserID    *uuid.UUID `db:"user_id"    json:"user_id,omitempty"`
	CreatedAt time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt time.Time  `db:"updated_at" json:"updated_at"`
}
