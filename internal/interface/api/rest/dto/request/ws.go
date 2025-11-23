package request

import (
	"github.com/google/uuid"
)

type WebSocketRequest struct {
	ChatID *uuid.UUID        `json:"chat_id,omitempty"`
	Flags  []string          `json:"flags,omitempty"`
	Prompt string            `json:"prompt,omitempty"`
	Vars   map[string]string `json:"vars,omitempty"`
	Tool   string            `json:"tool,omitempty"`
	Model  string            `json:"model,omitempty"`
	Media  []string          `json:"media,omitempty"`
}
