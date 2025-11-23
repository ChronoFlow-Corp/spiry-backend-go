package entities

import (
	"time"

	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/domain"
	"github.com/google/uuid"
)

// Plan is current user subscription.
type Plan struct {
	ID uuid.UUID

	// Quote will compute of plan created at and subscription Period.
	ModalitiesQuote ModalitiesQuote
	SubscriptionID  uuid.UUID
	Level           uint

	UserID    uuid.UUID
	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewPlan(quote ModalitiesQuote, level uint, subscriptionID uuid.UUID, userID uuid.UUID) *Plan {
	return &Plan{
		ID:              uuid.New(),
		ModalitiesQuote: quote,
		SubscriptionID:  subscriptionID,
		Level:           level,
		UserID:          userID,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now().Add(time.Millisecond),
	}
}

func (p *Plan) Validate() error {
	if p.SubscriptionID == uuid.Nil {
		return domain.NewValidationError(nil, "subscription_id", "subscription_id is required")
	}

	if p.ModalitiesQuote.ChattingQuote != nil && *p.ModalitiesQuote.ChattingQuote == 0 {
		return domain.NewValidationError(
			nil,
			"chatting_quote",
			"chatting_quote is required or be nil if no quote",
		)
	}

	if p.ModalitiesQuote.MediaQuote != nil && *p.ModalitiesQuote.MediaQuote == 0 {
		return domain.NewValidationError(
			nil,
			"media_quote",
			"media_quote is required or be nil if no quote",
		)
	}

	if p.ModalitiesQuote.TextContentQuote != nil && *p.ModalitiesQuote.TextContentQuote == 0 {
		return domain.NewValidationError(
			nil,
			"text_content_quote",
			"text_content_quote is required or be nil if no quote",
		)
	}

	if p.UserID == uuid.Nil {
		return domain.NewValidationError(nil, "user_id", "user_id is required")
	}

	if p.CreatedAt.After(p.UpdatedAt) {
		return domain.NewValidationError(nil, "created_at", "createdAt is in the future")
	}

	return nil
}
