package models

import (
	"time"

	"github.com/google/uuid"
)

type Tool struct {
	ID         uuid.UUID         `db:"id"`
	Name       string            `db:"name"`
	Modalities []string          `db:"modalities"`
	Settings   map[string]string `db:"settings"`
	Prompt     string            `db:"prompt"`
	MinLevel   int               `db:"min_level"`
	CreatedAt  time.Time         `db:"created_at"`
	UpdatedAt  time.Time         `db:"updated_at"`
}
