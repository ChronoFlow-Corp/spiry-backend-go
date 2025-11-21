package models

import (
	"time"

	"github.com/google/uuid"
)

type Command struct {
	ID             uuid.UUID         `db:"command_id"               json:"id,omitempty"`
	Prompt         string            `db:"command_prompt"           json:"prompt,omitempty"`
	Settings       map[string]string `db:"command_settings"         json:"settings,omitempty"`
	Flags          []string          `db:"command_flags"            json:"flags,omitempty"`
	ChatID         uuid.UUID         `db:"command_chat_id"          json:"chat_id,omitempty"`
	ToolID         uuid.UUID         `db:"command_tool_id"          json:"tool_id,omitempty"`
	ModelID        *uuid.UUID        `db:"command_model_id"         json:"model_id,omitempty"`
	Status         string            `db:"command_status"           json:"status,omitempty"`
	UserID         uuid.UUID         `db:"command_user_id"          json:"user_id,omitempty"`
	UnloggedUserID uuid.UUID         `db:"command_unlogged_user_id" json:"unlogged_user_id,omitempty"`
	CreatedAt      time.Time         `db:"command_created_at"       json:"created_at"`
	UpdatedAt      time.Time         `db:"command_updated_at"       json:"updated_at"`
}

type Tool struct {
	ID         uuid.UUID         `db:"tool_id"`
	Name       string            `db:"tool_name"`
	Modalities []string          `db:"tool_modalities"`
	Settings   map[string]string `db:"tool_settings"`
	Prompt     string            `db:"tool_prompt"`
	MinLevel   int               `db:"tool_min_level"`
	CreatedAt  time.Time         `db:"tool_created_at"`
	UpdatedAt  time.Time         `db:"tool_updated_at"`
}

type CommandMedia struct {
	ID             uuid.UUID `db:"command_media_id"               json:"id,omitempty"`
	Name           string    `db:"command_media_name"             json:"name"`
	Type           string    `db:"command_media_type"             json:"type"`
	URL            string    `db:"command_media_url"              json:"url"`
	Size           int       `db:"command_media_size"             json:"size"`
	CommandID      uuid.UUID `db:"command_media_command_id"       json:"command_id,omitempty"`
	UserID         uuid.UUID `db:"command_media_user_id"          json:"user_id,omitempty"`
	UnloggedUserID uuid.UUID `db:"command_media_unlogged_user_id" json:"unlogged_user_id,omitempty"`
	CreatedAt      time.Time `db:"command_media_created_at"       json:"created_at"`
	UpdatedAt      time.Time `db:"command_media_updated_at"       json:"updated_at"`
}
