package auth

import (
	"context"

	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/application/command"
	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/application/result"
)

type Service interface {
	Login(ctx context.Context, cm command.Login) (result.Login, error)
	GetAuthURI(device string) string
	Refresh(ctx context.Context, cm command.Refresh) (result.Refresh, error)
	LogOut(ctx context.Context, cm command.Logout) error
	UserInfo(ctx context.Context) (result.UserInfo, error)
}
