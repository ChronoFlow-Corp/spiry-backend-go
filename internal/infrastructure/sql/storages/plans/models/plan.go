package models

import (
	"time"

	"github.com/google/uuid"
)

type Plan struct {
	ID             uuid.UUID `db:"id"`
	Name           string    `db:"name"`
	Quote          []byte    `db:"quote"`
	SubscriptionID uuid.UUID `db:"subscription_id"`
	Level          uint      `db:"level"`
	UserID         uuid.UUID `db:"user_id"`
	CreatedAt      time.Time `db:"created_at"`
	UpdatedAt      time.Time `db:"updated_at"`
}

type Quote struct {
	ToolLimits       []ToolLimit  `json:"tool_limits"`
	MediaLimit       []MediaLimit `json:"media_limit"`
	FlagLimits       []FlagLimit  `json:"flag_limits"`
	ResetQuotePeriod string       `json:"reset_quote_period"`
}

type ToolLimit struct {
	ID            uuid.UUID       `json:"id"`
	SettingsLimit map[string]bool `json:"settings_limit"`
	Usage         int             `json:"usage"`
}

type MediaLimit struct {
	Type     string `json:"type"`
	Upload   int    `json:"upload"`
	Generate int    `json:"generate"`
	Size     int    `json:"size"`
}

type FlagLimit struct {
	Name string `json:"name"`
}
