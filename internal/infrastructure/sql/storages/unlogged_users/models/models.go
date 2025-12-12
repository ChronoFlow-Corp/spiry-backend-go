package models

import (
	"time"

	"github.com/google/uuid"
)

type UnloggedUser struct {
	ID        uuid.UUID `db:"id"`
	IP        string    `db:"ip_address"`
	PlanID    uuid.UUID `db:"plan_id"`
	CreatedAt time.Time `db:"created_at"`
}
