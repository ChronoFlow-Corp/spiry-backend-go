package unloggedusers

import (
	"context"

	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/domain/entities"
)

type UnloggedUserStorage interface {
	Create(ctx context.Context, u *entities.UnloggedUser) error
	GetByIP(ctx context.Context, ip string) (*entities.UnloggedUser, error)
}
