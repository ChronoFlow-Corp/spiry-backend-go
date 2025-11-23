package command

import "github.com/google/uuid"

type UpdateTitle struct {
	ChatID uuid.UUID
	Title  string
}
