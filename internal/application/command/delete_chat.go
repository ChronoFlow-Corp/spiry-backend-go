package command

import "github.com/google/uuid"

type DeleteChat struct {
	ChatID uuid.UUID
}
