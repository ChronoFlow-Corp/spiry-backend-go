package command

import "github.com/google/uuid"

type Logout struct {
	SessionID uuid.UUID
	UserID    uuid.UUID
}
