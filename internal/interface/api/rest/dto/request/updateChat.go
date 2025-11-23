package request

import "github.com/google/uuid"

type UpdateChat struct {
	ChatID   uuid.UUID `json:"chat_id"`
	NewTitle string    `json:"new_title"`
}
