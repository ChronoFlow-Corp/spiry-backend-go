package models

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type Subscription struct {
	ID        uuid.UUID      `db:"id"`
	Name      string         `db:"name"`
	Quote     []byte         `db:"quote"`
	Period    string         `db:"period"`
	Price     sql.NullString `db:"price"`
	Level     uint           `db:"level"`
	CreatedAt time.Time      `db:"created_at"`
	UpdatedAt time.Time      `db:"updated_at"`
}

type Quote struct {
	ToolLimits       []ToolLimit  `json:"tool_limits,omitempty"`
	MediaLimit       []MediaLimit `json:"media_limit,omitempty"`
	FlagLimits       []FlagLimit  `json:"flag_limits,omitempty"`
	ResetQuotePeriod string       `json:"reset_quote_period,omitempty"`
}

type ToolLimit struct {
	ID            uuid.UUID       `json:"id"`
	SettingsLimit map[string]bool `json:"settings_limit,omitempty"`
	Usage         int             `json:"usage,omitempty"`
}

type MediaLimit struct {
	Type     string `json:"type,omitempty"`
	Upload   int    `json:"upload,omitempty"`
	Generate int    `json:"generate,omitempty"`
	Size     uint64 `json:"size,omitempty"`
}

type FlagLimit struct {
	Name string `json:"name,omitempty"`
}
