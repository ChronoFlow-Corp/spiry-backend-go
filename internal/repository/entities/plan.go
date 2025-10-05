package entities

import (
	"time"

	"github.com/google/uuid"
)

const (
	FreePlanName = "free"
	ProPlanName  = "pro"
)

type Plan struct {
	ID        uuid.UUID
	Name      string
	Limit     *int
	Price     string
	Features  []string
	End       *time.Time
	UserID    uuid.UUID
	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewFreePlan() Plan {
	limit := 5
	end := time.Now().Add(time.Hour * 24)
	return Plan{
		ID:    uuid.New(),
		Name:  FreePlanName,
		Limit: &limit,
		Price: "0",
		End:   &end,
	}
}
