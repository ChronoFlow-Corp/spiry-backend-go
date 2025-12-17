package entities

import (
	"time"

	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/domain"
	"github.com/google/uuid"
)

const (
	maxLenPrompt   = 2048
	maxLenFlags    = 30
	maxLenSettings = 128
)

const (
	StatusInProgress = "in_progress"
	StatusDone       = "done"
	StatusError      = "error"
)

const (
	FlagWebSearch = "web_search"
)

// Command is entity of command to llm.
type Command struct {
	ID       uuid.UUID
	Prompt   string
	Settings map[string]string

	// Flags is "web search" etc...
	Flags  []string
	ChatID uuid.UUID
	ToolID *uuid.UUID

	// ModelID can be nil if user don't set model
	ModelID   *uuid.UUID
	Status    string
	UserID    *uuid.UUID
	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewCommand(
	prompt string,
	settings map[string]string,
	flags []string,
	status string,
	modelID *uuid.UUID,
	chatID uuid.UUID,
	toolID *uuid.UUID,
	userID *uuid.UUID,
) *Command {
	return &Command{
		ID:        uuid.New(),
		Prompt:    prompt,
		Settings:  settings,
		Flags:     flags,
		Status:    status,
		ChatID:    chatID,
		ModelID:   modelID,
		ToolID:    toolID,
		UserID:    userID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func (c *Command) Validate() error {
	if len(c.Prompt) > maxLenPrompt {
		return domain.NewValidationError(nil, "prompt", "prompt is too long")
	}

	if len(c.Settings) > maxLenSettings {
		return domain.NewValidationError(nil, "settings", "settings is too long")
	}

	if len(c.Flags) > maxLenFlags {
		return domain.NewValidationError(nil, "flags", "flags is too long")
	}

	if c.ChatID == uuid.Nil {
		return domain.NewValidationError(nil, "chat_id", "chat_id is required")
	}

	if c.Status != StatusDone && c.Status != StatusError && c.Status != StatusInProgress {
		return domain.NewValidationError(nil, "status", "status is invalid")
	}

	for _, flag := range c.Flags {
		switch flag {
		case FlagWebSearch:
		default:
			return domain.NewValidationError(nil, "flags", "flag is invalid")
		}
	}

	if c.CreatedAt.After(c.UpdatedAt) {
		return domain.NewValidationError(nil, "created_at", "createdAt is in the future")
	}

	return nil
}
