package query

import "github.com/google/uuid"

type GetChats struct {
	ChatID *uuid.UUID
}
