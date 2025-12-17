package chatting

import (
	"context"
	"errors"
	"fmt"
	"net/url"

	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/application/service/chatting"
	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/domain"
	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/domain/aggregates"
	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/domain/entities"
	models2 "github.com/ChronoFlow-Corp/spiry-backend-go/internal/domain/models"
	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/infrastructure/sql/storages/chats"
	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/infrastructure/sql/storages/commands"
	commandsmedias "github.com/ChronoFlow-Corp/spiry-backend-go/internal/infrastructure/sql/storages/commands_medias"
	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/infrastructure/sql/storages/models"
	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/infrastructure/sql/storages/plans"
	resultmedias "github.com/ChronoFlow-Corp/spiry-backend-go/internal/infrastructure/sql/storages/result_medias"
	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/infrastructure/sql/storages/results"
	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/infrastructure/sql/storages/subscriptions"
	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/infrastructure/sql/storages/tools"
	"github.com/avito-tech/go-transaction-manager/trm/v2/manager"
	"github.com/google/uuid"
)

type Repository struct {
	chat         chats.ChatStorage
	command      commands.CommandStorage
	model        models.ModelStorage
	commandMedia commandsmedias.CommandMediaStorage
	result       results.ResultStorage
	resultMedia  resultmedias.ResultMediaStorage
	tool         tools.ToolStorage
	plan         plans.PlanStorage
	sub          subscriptions.SubscriptionStorage
	manager      *manager.Manager
}

var _ chatting.UseCaseRepository = (*Repository)(nil)

func NewRepository(
	chat chats.ChatStorage,
	command commands.CommandStorage,
	commandMedia commandsmedias.CommandMediaStorage,
	result results.ResultStorage,
	resultMedia resultmedias.ResultMediaStorage,
	model models.ModelStorage,
	tool tools.ToolStorage,
	plan plans.PlanStorage,
	sub subscriptions.SubscriptionStorage,
	manager *manager.Manager,
) *Repository {
	return &Repository{
		chat:         chat,
		command:      command,
		model:        model,
		commandMedia: commandMedia,
		result:       result,
		resultMedia:  resultMedia,
		tool:         tool,
		plan:         plan,
		sub:          sub,
		manager:      manager,
	}
}

func (r *Repository) SaveCommand(ctx context.Context, cm *aggregates.Command) error {
	const op = "repository.chatting.SaveCommand"

	err := r.manager.Do(ctx, func(ctx context.Context) error {
		_, err := r.command.GetByID(ctx, cm.ID)
		switch {
		case errors.Is(err, commands.ErrNotFound):
			err = r.command.Create(ctx, *cm.Command)
			if err != nil {
				return err
			}
		case err != nil && !errors.Is(err, commands.ErrNotFound):
			return err
		default:
			err = r.command.Update(ctx, *cm.Command)
			if err != nil {
				return err
			}
		}

		for _, m := range cm.Medias {
			err = r.commandMedia.Update(ctx, *m)
			if err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (r *Repository) GetChatByID(ctx context.Context, chatID uuid.UUID) (*aggregates.Chat, error) {
	const op = "repository.chatting.GetChatByID"

	chat, err := r.chat.GetByIDWithCommandsResults(ctx, chatID)
	if err != nil {
		if errors.Is(err, chats.ErrNotFound) {
			return nil, domain.NewNotFound(err, "", "chatID", chatID.String())
		}

		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return chat, nil
}

func (r *Repository) GetChatByUserID(ctx context.Context) ([]*aggregates.Chat, error) {
	const op = "repository.chatting.GetChatByUserID"

	chat, err := r.chat.GetAll(ctx)
	if err != nil {
		if errors.Is(err, chats.ErrNotFound) {
			return nil, domain.NewNotFound(err, "chat not found", "userID", "")
		}

		return nil, fmt.Errorf("%s: %w", op, err)
	}

	ch := make([]*aggregates.Chat, 0, len(chat))
	for _, m := range chat {
		ch = append(ch, &aggregates.Chat{
			Chat: m,
		})
	}

	return ch, nil
}

func (r *Repository) DeleteChatByID(ctx context.Context, chatID uuid.UUID) error {
	const op = "repository.chatting.DeleteChatByID"

	err := r.chat.Delete(ctx, chatID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (r *Repository) SaveChat(ctx context.Context, chat *aggregates.Chat) error {
	const op = "repository.chatting.SaveChat"

	err := r.manager.Do(ctx, func(ctx context.Context) error {
		_, err := r.chat.GetByID(ctx, chat.ID)
		switch {
		case errors.Is(err, chats.ErrNotFound):
			err = r.chat.Create(ctx, *chat.Chat)
			if err != nil {
				return err
			}
		case err != nil && !errors.Is(err, chats.ErrNotFound):
			return err
		default:
			err = r.chat.Update(ctx, *chat.Chat)
			if err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (r *Repository) SaveResult(ctx context.Context, result *aggregates.Result) error {
	const op = "repository.chatting.SaveResult"

	err := r.manager.Do(ctx, func(ctx context.Context) error {
		_, err := r.command.GetByID(ctx, result.ID)
		switch {
		case errors.Is(err, commands.ErrNotFound):
			err = r.result.Create(ctx, *result.Result)
			if err != nil {
				return err
			}
		case err != nil && !errors.Is(err, commands.ErrNotFound):
			return err
		default:
			return err
		}

		for _, m := range result.Medias {
			err = r.resultMedia.Update(ctx, *m)
			if err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (r *Repository) GetModels(ctx context.Context) ([]*entities.Model, error) {
	const op = "repository.chatting.GetModels"

	modelsList, err := r.model.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	m := make([]*entities.Model, 0, len(modelsList))
	for _, model := range modelsList {
		m = append(m, &model)
	}

	return m, nil
}

func (r *Repository) GetMediaByUrls(
	_ context.Context,
	_ []*url.URL,
) ([]*entities.CommandMedia, error) {
	const _ = "repository.chatting.GetMediaByUrls"

	return nil, nil
}

func (r *Repository) GetModelByName(ctx context.Context, name string) (*entities.Model, error) {
	const op = "repository.chatting.GetModelByName"

	model, err := r.model.GetByName(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return model, nil
}

func (r *Repository) GetTools(ctx context.Context) ([]*entities.Tool, error) {
	const op = "repository.chatting.GetTools"

	t, err := r.tool.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return t, nil
}

func (r *Repository) GetAllowedTools(ctx context.Context) ([]*entities.Tool, error) {
	const op = "repository.chatting.GetAllowedTools"

	t, err := r.tool.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	allowed := make([]*entities.Tool, 0, len(t))

	id, ok := models2.GetUserIDFromCtx(ctx)
	if !ok {
		for _, m := range t {
			if m.MinLevel == 0 {
				allowed = append(allowed, m)
			}
		}

		return allowed, nil
	}

	p, err := r.plan.GetByUserID(ctx, *id)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	sub, err := r.sub.GetByID(ctx, p.SubscriptionID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	p.Level = sub.Level

	for _, m := range t {
		for _, tl := range p.Quote.ToolLimits {
			if tl.Usage != 0 && m.MinLevel <= p.Level && tl.ID == m.ID {
				allowed = append(allowed, m)
			}
		}
	}

	if len(allowed) == 0 {
		return nil, domain.NewForbidden(
			domain.ErrZeroAllowedTools,
			fmt.Sprintf("%s: zero allowed tools", op),
			"",
			"",
		)
	}

	return allowed, nil
}

func (r *Repository) GetAllowedModels(ctx context.Context) ([]*entities.Model, error) {
	const op = "repository.chatting.GetAllowedModels"

	mList, err := r.model.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	allowed := make([]*entities.Model, 0, len(mList))

	id, ok := models2.GetUserIDFromCtx(ctx)
	if !ok {
		for _, m := range mList {
			if m.MinLevel == 0 {
				allowed = append(allowed, &m)
			}
		}

		return allowed, nil
	}

	p, err := r.plan.GetByUserID(ctx, *id)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	sub, err := r.sub.GetByID(ctx, p.SubscriptionID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	p.Level = sub.Level

	for _, m := range mList {
		if p.Level >= m.MinLevel {
			allowed = append(allowed, &m)
		}
	}

	return allowed, nil
}

func (r *Repository) GetToolByName(ctx context.Context, name string) (*entities.Tool, error) {
	const op = "repository.chatting.GetToolByName"

	t, err := r.tool.GetByName(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return t, nil
}
