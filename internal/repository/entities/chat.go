package entities

import (
	"time"

	"github.com/google/uuid"
)

type Chat struct {
	ID        uuid.UUID
	Title     string
	ToolID uuid.UUID
	UserID    uuid.UUID
	CreatedAt time.Time
	UpdatedAt time.Time
	Messages []Message
}

func NewChat(ID uuid.UUID, title string, tool uuid.UUID, userID uuid.UUID, messages []Message) Chat {
	return Chat{ID: ID, Title: title, ToolID: tool, UserID: userID, Messages: messages}
}
