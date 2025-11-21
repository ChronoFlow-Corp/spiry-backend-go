package models

import (
	"time"

	"github.com/google/uuid"
)

type Result struct {
	ID           uuid.UUID  `db:"result_id"             json:"id,omitempty"`
	OpenRouterID string     `db:"result_open_router_id" json:"open_router_id,omitempty"`
	Text         string     `db:"result_text"           json:"text,omitempty"`
	CommandID    uuid.UUID  `db:"result_command_id"     json:"command_id,omitempty"`
	ToolID       uuid.UUID  `db:"result_tool_id"        json:"tool_id,omitempty"`
	ChatID       uuid.UUID  `db:"result_chat_id"        json:"chat_id,omitempty"`
	UserID       *uuid.UUID `db:"result_user_id"        json:"user_id,omitempty"`
	ModelID      uuid.UUID  `db:"model_id"              json:"model_id,omitempty"`
	CreatedAt    time.Time  `db:"result_created_at"     json:"created_at"`
	UpdatedAt    time.Time  `db:"result_updated_at"     json:"updated_at"`
}

type ResultMedia struct {
	ID        uuid.UUID  `db:"result_media_id"         json:"id,omitempty"`
	Name      string     `db:"result_media_name"       json:"name,omitempty"`
	Type      string     `db:"result_media_type"       json:"type,omitempty"`
	URL       string     `db:"result_media_url"        json:"url,omitempty"`
	Size      int        `db:"result_media_size"       json:"size,omitempty"`
	ResultID  uuid.UUID  `db:"result_media_result_id"  json:"result_id,omitempty"`
	UserID    *uuid.UUID `db:"result_media_user_id"    json:"user_id,omitempty"`
	CreatedAt time.Time  `db:"result_media_created_at" json:"created_at"`
	UpdatedAt time.Time  `db:"result_media_updated_at" json:"updated_at"`
}
