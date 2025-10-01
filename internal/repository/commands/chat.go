package commands

import "github.com/google/uuid"

type DeleteChat struct {
	ChatID uuid.UUID
}

type UpdateChatTitle struct {
	ChatID uuid.UUID
	NewTitle string
}
