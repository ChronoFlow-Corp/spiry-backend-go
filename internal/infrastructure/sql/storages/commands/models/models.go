package models

import (
	"time"

	"github.com/google/uuid"
)

type Command struct {
	ID        uuid.UUID         `db:"id"         json:"id,omitempty"`
	Prompt    string            `db:"prompt"     json:"prompt,omitempty"`
	Settings  map[string]string `db:"settings"   json:"settings,omitempty"`
	Flags     []string          `db:"flags"      json:"flags,omitempty"`
	ChatID    uuid.UUID         `db:"chat_id"    json:"chat_id,omitempty"`
	ToolID    *uuid.UUID        `db:"tool_id"    json:"tool_id,omitempty"`
	ModelID   *uuid.UUID        `db:"model_id"   json:"model_id,omitempty"`
	Status    string            `db:"status"     json:"status,omitempty"`
	UserID    *uuid.UUID        `db:"user_id"    json:"user_id,omitempty"`
	CreatedAt time.Time         `db:"created_at" json:"created_at"`
	UpdatedAt time.Time         `db:"updated_at" json:"updated_at"`
}
