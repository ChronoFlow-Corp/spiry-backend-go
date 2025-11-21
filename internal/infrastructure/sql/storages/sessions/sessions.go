package sessions

import (
	"context"

	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/domain/entities"
	"github.com/google/uuid"
)

type SessionStorage interface {
	Create(ctx context.Context, session entities.Session) error
	GetByToken(ctx context.Context, t string) (*entities.Session, error)
	GetByID(ctx context.Context, i uuid.UUID) (*entities.Session, error)
	Update(ctx context.Context, session entities.Session) error
	GetByUserID(ctx context.Context, u uuid.UUID) ([]*entities.Session, error)
}
