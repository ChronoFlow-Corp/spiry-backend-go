package entities

import (
	"time"

	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/domain"
	"github.com/google/uuid"
)

const (
	maxLenToolName     = 64
	maxLenToolSettings = 64
	maxLenPromptTool   = 4096
)

type Modality string

// Tool is pre prompt for llm.
// Model will choose of subscription level, modalities, priority.
type Tool struct {
	ID         uuid.UUID
	Name       string
	Modalities []Modality
	Settings   map[string]string
	Prompt     string
	MinLevel   uint
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

func NewTool(name string, settings map[string]string, prompt string, minLevel uint) *Tool {
	return &Tool{
		ID:        uuid.New(),
		Name:      name,
		Settings:  settings,
		MinLevel:  minLevel,
		Prompt:    prompt,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func (t *Tool) Validate() error {
	if t.Name == "" {
		return domain.NewValidationError(nil, "name", "name is required")
	}

	if len(t.Name) > maxLenToolName {
		return domain.NewValidationError(nil, "name", "name is too long")
	}

	if len(t.Settings) > maxLenToolSettings {
		return domain.NewValidationError(nil, "settings", "settings is too long")
	}

	if len(t.Prompt) > maxLenPromptTool {
		return domain.NewValidationError(nil, "prompt", "prompt is too long")
	}

	if t.CreatedAt.After(t.UpdatedAt) {
		return domain.NewValidationError(nil, "created_at", "createdAt is in the future")
	}

	return nil
}
