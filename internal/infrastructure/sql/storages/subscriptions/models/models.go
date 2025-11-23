package models

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type Subscription struct {
	ID              uuid.UUID      `db:"id"`
	Name            string         `db:"name"`
	ModalitiesQuote []byte         `db:"modalities_quote"`
	Period          string         `db:"period"`
	Price           sql.NullString `db:"price"`
	Level           int            `db:"level"`
	CreatedAt       time.Time      `db:"created_at"`
	UpdatedAt       time.Time      `db:"updated_at"`
}

type ModalitiesQuote struct {
	MediaQuote       *uint `json:"media_quote,omitempty"`
	TextContentQuote *uint `json:"text_content_quote,omitempty"`
	ChattingQuote    *uint `json:"chatting_quote,omitempty"`
}
