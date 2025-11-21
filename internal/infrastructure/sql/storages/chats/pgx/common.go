package pgx

import (
	"context"

	domainModels "github.com/ChronoFlow-Corp/spiry-backend-go/internal/domain/models"
	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
)

func addUserIDWhere(ctx context.Context, builder any) {
	uID, ok := ctx.Value(domainModels.UserIDCtxKey{}).(uuid.UUID)

	switch b := builder.(type) {
	case squirrel.DeleteBuilder:
		if ok {
			b.Where(squirrel.Eq{columns[userID]: uID})
		}
	case squirrel.UpdateBuilder:
		if ok {
			b.Where(squirrel.Eq{columns[userID]: uID})
		}
	case squirrel.SelectBuilder:
		if ok {
			b.Where(squirrel.Eq{columns[userID]: uID})
		}
	}
}

func getUserId(ctx context.Context) (uuid.UUID, bool) {
	uID, ok := ctx.Value(domainModels.UserIDCtxKey{}).(uuid.UUID)
	return uID, ok
}
