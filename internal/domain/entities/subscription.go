package entities

import (
	"time"

	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/domain"
	"github.com/google/uuid"
)

type Period string

// Time to expires subscription.
const (
	Day   Period = "day"
	Month Period = "month"
	Year  Period = "year"
)

const (
	maxLenSubscriptionName = 64
)

// Subscription is existing plan configuration.
type Subscription struct {
	ID              uuid.UUID
	Name            string
	ModalitiesQuote ModalitiesQuote
	Period          Period
	Price           string
	Level           uint
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type ModalitiesQuote struct {
	MediaQuote       *uint
	TextContentQuote *uint
	ChattingQuote    *uint
}

func NewSubscription(name string, period Period, quote ModalitiesQuote, price string, level uint) *Subscription {
	return &Subscription{
		ID:              uuid.New(),
		Name:            name,
		ModalitiesQuote: quote,
		Period:          period,
		Price:           price,
		Level:           level,
	}
}

func (s *Subscription) Validate() error {
	if s.Name == "" {
		return domain.NewValidationError(nil, "name", "name is required")
	}

	if len(s.Name) > maxLenSubscriptionName {
		return domain.NewValidationError(nil, "name", "name is too long")
	}

	if s.ModalitiesQuote.ChattingQuote != nil && *s.ModalitiesQuote.ChattingQuote == 0 {
		return domain.NewValidationError(
			nil,
			"chatting_quote",
			"chatting_quote is required or be nil if no quote",
		)
	}

	if s.ModalitiesQuote.MediaQuote != nil && *s.ModalitiesQuote.MediaQuote == 0 {
		return domain.NewValidationError(
			nil,
			"media_quote",
			"media_quote is required or be nil if no quote",
		)
	}

	if s.ModalitiesQuote.TextContentQuote != nil && *s.ModalitiesQuote.TextContentQuote == 0 {
		return domain.NewValidationError(
			nil,
			"text_content_quote",
			"text_content_quote is required or be nil if no quote",
		)
	}

	switch s.Period {
	case Month:
	case Year:
	default:
		return domain.NewValidationError(nil, "Period", "invalid Period")
	}

	if s.CreatedAt.After(s.UpdatedAt) {
		return domain.NewValidationError(nil, "created_at", "createdAt is in the future")
	}

	return nil
}
