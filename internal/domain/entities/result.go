package entities

import (
	"time"

	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/domain"
	"github.com/google/uuid"
)

// Result generation of llm.
type Result struct {
	ID           uuid.UUID
	Text         string
	OpenRouterID string
	CommandID    uuid.UUID
	ToolID       uuid.UUID
	ChatID       uuid.UUID
	UserID       *uuid.UUID
	ModelID      uuid.UUID
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func NewResult(
	text string,
	openRouterID string,
	commandID uuid.UUID,
	toolID uuid.UUID,
	chatID uuid.UUID,
	modelID uuid.UUID,
	userID *uuid.UUID) *Result {
	return &Result{
		ID:           uuid.New(),
		Text:         text,
		OpenRouterID: openRouterID,
		CommandID:    commandID,
		ToolID:       toolID,
		ModelID:      modelID,
		ChatID:       chatID,
		UserID:       userID,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
}

func (r *Result) Validate() error {
	if r.ChatID == uuid.Nil {
		return domain.NewValidationError(nil, "chat_id", "chat_id is required")
	}

	if r.ToolID == uuid.Nil {
		return domain.NewValidationError(nil, "tool_id", "tool_id is required")
	}

	if r.OpenRouterID == "" {
		return domain.NewValidationError(nil, "open_router_id", "open_router_id is required")
	}

	if r.CreatedAt.After(r.UpdatedAt) {
		return domain.NewValidationError(nil, "created_at", "createdAt is in the future")
	}

	return nil
}
