package commands

import (
	"io"

	"github.com/google/uuid"
)

//TODO: tool type is attribute of chat, not message
type SendMessage struct {
	ChatID   *uuid.UUID
	Role string
	Vars map[string]string
	Flags []string
	Tool string
	Media io.ReadCloser
	Network string
	Language string
}

type GetMessages struct {
	ChatID uuid.UUID
	Limit  int
	Offset int
}

type DeleteMessage struct {
	MessageID uuid.UUID
}