package response

import (
	"time"

	"github.com/google/uuid"
)

type UserInfo struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	AvatarURL string    `json:"avatar_url"`
	Plan      Plan      `json:"plan"`
	Theme     string    `json:"theme"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Plan struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Price     string    `json:"price"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
