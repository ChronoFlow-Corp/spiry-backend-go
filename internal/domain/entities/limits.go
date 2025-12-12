package entities

import "github.com/google/uuid"

type Quote struct {
	ToolLimits       []ToolLimit
	MediaLimit       []MediaLimit
	FlagLimits       []FlagLimit
	ResetQuotePeriod Period
}

type ToolLimit struct {
	ID uuid.UUID

	// SettingsLimit indicates whether the limit is related to user settings.
	SettingsLimit map[string]bool

	// Usage indicates the current usage count for the tool limit.
	Usage int
}

type MediaLimit struct {
	Type     string
	Upload   int
	Generate int
	Size     uint64
}

type FlagLimit struct {
	Name string
}
