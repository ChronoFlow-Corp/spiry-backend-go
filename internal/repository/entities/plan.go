package entities

import (
	"time"

	"github.com/google/uuid"
)

const (
	FreePlanName  = "free"
	ProPlanName = "pro"
)

type Plan struct {
	ID     uuid.UUID
	Name   string
	Limit  *int
	Price string
	Features []string
	End    *time.Time
	UserID uuid.UUID
	CreatedAt time.Time
	UpdatedAt time.Time
}
