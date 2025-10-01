package entities

import (
	"time"

	"github.com/google/uuid"
)

type Tool struct{
	ID uuid.UUID
	Name string
	Model string
	Prompt string
	Vars map[string]string
	CreatedAt time.Time
	UpdatedAt time.Time
}
