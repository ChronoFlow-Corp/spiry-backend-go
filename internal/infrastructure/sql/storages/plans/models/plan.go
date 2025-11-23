package models

import (
	"time"

	"github.com/google/uuid"
)

type Plan struct {
	ID              uuid.UUID `db:"id"`
	Name            string    `db:"name"`
	ModalitiesQuote []byte    `db:"modalities_quote"`
	SubscriptionID  uuid.UUID `db:"subscription_id"`
	Level           uint      `db:"level"`
	UserID          uuid.UUID `db:"user_id"`
	CreatedAt       time.Time `db:"created_at"`
	UpdatedAt       time.Time `db:"updated_at"`
}

type ModalitiesQuote struct {
	MediaQuote       *uint `json:"media_quote,omitempty"`
	TextContentQuote *uint `json:"text_content_quote,omitempty"`
	ChattingQuote    *uint `json:"chatting_quote,omitempty"`
}
