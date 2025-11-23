package models

import (
	"time"

	"github.com/google/uuid"
)

type ResultMedia struct {
	ID        uuid.UUID  `db:"id"         json:"id,omitempty"`
	Name      string     `db:"name"       json:"name,omitempty"`
	Type      string     `db:"type"       json:"type,omitempty"`
	URL       string     `db:"url"        json:"url,omitempty"`
	Size      int        `db:"size"       json:"size,omitempty"`
	ResultID  uuid.UUID  `db:"result_id"  json:"result_id,omitempty"`
	UserID    *uuid.UUID `db:"user_id"    json:"user_id,omitempty"`
	CreatedAt time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt time.Time  `db:"updated_at" json:"updated_at"`
}
