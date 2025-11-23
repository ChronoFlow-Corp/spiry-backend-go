package result

import (
	"time"

	"github.com/google/uuid"
)

type UserInfo struct {
	ID        uuid.UUID
	Name      string
	Email     string
	AvatarURL string
	Plan      Plan
	Theme     string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Plan struct {
	ID        uuid.UUID
	Name      string
	Price     string
	CreatedAt time.Time
	UpdatedAt time.Time
}
