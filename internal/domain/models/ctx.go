package models

import (
	"context"

	"github.com/google/uuid"
)

type UserIDCtxKey struct{}

type SessionIDCtxKey struct{}

func GetUserIDFromCtx(ctx context.Context) (*uuid.UUID, bool) {
	id, ok := ctx.Value(UserIDCtxKey{}).(uuid.UUID)
	if !ok {
		return nil, false
	}

	return &id, ok
}
