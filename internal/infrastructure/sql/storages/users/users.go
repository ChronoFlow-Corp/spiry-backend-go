package users

import (
	"context"

	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/domain/entities"
	"github.com/google/uuid"
)

type UserStorage interface {
	Create(ctx context.Context, user entities.User) error
	GetByID(ctx context.Context, userId uuid.UUID) (*entities.User, error)
	GetByEmail(ctx context.Context, e string) (*entities.User, error)
	Update(ctx context.Context, user entities.User) error
}
