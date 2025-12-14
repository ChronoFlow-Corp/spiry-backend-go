package response

import (
	"time"

	"github.com/google/uuid"
)

const (
	UserRole      = "user"
	AssistantRole = "assistant"
)

type Chat struct {
	ID        uuid.UUID `json:"id"`
	Title     string    `json:"title"`
	Messages  []Message `json:"messages,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Message struct {
	ID        uuid.UUID         `json:"id"`
	Text      string            `json:"content"`
	Settings  map[string]string `json:"settings,omitempty"`
	Flags     []string          `json:"flags,omitempty"`
	Status    *string           `json:"status,omitempty"`
	Tool      *Tool             `json:"tool,omitempty"`
	Model     *Model            `json:"model,omitempty"`
	Role      string            `json:"role,omitempty"`
	CreatedAt time.Time         `json:"created_at"`
}

type Tool struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

type Model struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}
