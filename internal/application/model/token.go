package model

import (
	"time"

	"github.com/google/uuid"
)

type Token struct {
	Token    string
	ExpireAt time.Time
}

type ParsedToken struct {
	UserID    uuid.UUID
	SessionID uuid.UUID
	ExpireAt  time.Time
}
