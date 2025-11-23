package chatting

import (
	"context"
	"net/url"

	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/domain/aggregates"
	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/domain/entities"
	"github.com/google/uuid"
)

type UseCaseRepository interface {
	SaveCommand(ctx context.Context, command *aggregates.Command) error
	GetChatByID(ctx context.Context, id uuid.UUID) (*aggregates.Chat, error)
	GetChatByUserID(ctx context.Context) ([]*aggregates.Chat, error)
	DeleteChatByID(ctx context.Context, id uuid.UUID) error
	SaveChat(ctx context.Context, chat *aggregates.Chat) error
	GetTools(ctx context.Context) ([]*entities.Tool, error)
	GetToolByName(ctx context.Context, name string) (*entities.Tool, error)
	SaveResult(ctx context.Context, result *aggregates.Result) error
	GetMediaByUrls(ctx context.Context, urls []*url.URL) ([]*entities.CommandMedia, error)
	GetModels(ctx context.Context) ([]*entities.Model, error)
	GetModelByName(ctx context.Context, name string) (*entities.Model, error)
	GetAllowedModels(ctx context.Context) ([]*entities.Model, error)
	GetAllowedTools(ctx context.Context) ([]*entities.Tool, error)
}

type AuthRepository interface {
	SaveUser(ctx context.Context, user *aggregates.User) error
	GetUserByEmail(ctx context.Context, email string) (*aggregates.User, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (*aggregates.User, error)
}
