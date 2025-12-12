package entities

import (
	"time"

	"github.com/google/uuid"
)

type UnloggedUser struct {
	ID        uuid.UUID
	IP        string
	PlanID    uuid.UUID
	CreatedAt time.Time
}

func NewUnloggedUser(ip string, planID uuid.UUID) *UnloggedUser {
	return &UnloggedUser{
		ID:        uuid.New(),
		IP:        ip,
		PlanID:    planID,
		CreatedAt: time.Now().UTC(),
	}
}
