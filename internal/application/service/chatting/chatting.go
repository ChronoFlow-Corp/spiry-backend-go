package chatting

import (
	"context"
	"errors"
	"fmt"

	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/application/command"
	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/application/pkg/event"
	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/application/query"
	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/application/result"
	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/domain/entities"
	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/domain/models"
	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/domain/service"
	"github.com/google/uuid"
)

type Service interface {
	GetChats(ctx context.Context, q query.GetChats) ([]result.GetChats, error)
	UpdateTitle(ctx context.Context, cm command.UpdateTitle) error
	DeleteChat(ctx context.Context, cm command.DeleteChat) error
	ExecuteCommand(ctx context.Context, cm command.Execute) (*event.Manager, error)
}

var _ Service = (*Chatting)(nil)

type Chatting struct {
	repo           UseCaseRepository
	authRepo       AuthRepository
	chattingDomain *service.Chatting
}

func NewChatting(
	repo UseCaseRepository,
	authRepo AuthRepository,
	chattingDomain *service.Chatting) *Chatting {
	return &Chatting{
		chattingDomain: chattingDomain,
		repo:           repo,
		authRepo:       authRepo,
	}
}

func (c *Chatting) GetChats(ctx context.Context, q query.GetChats) ([]result.GetChats, error) {
	const op = "application.service.Chatting.GetChats"

	if q.ChatID == nil {
		chats, err := c.repo.GetChatByUserID(ctx)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		res := make([]result.GetChats, 0, len(chats))
		for _, chat := range chats {
			res = append(res, result.GetChats{
				ID:        chat.ID,
				Title:     chat.Title,
				CreatedAt: chat.CreatedAt,
				UpdatedAt: chat.UpdatedAt,
			})
		}

		return res, nil
	}

	chat, err := c.repo.GetChatByID(ctx, *q.ChatID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	res := result.GetChats{
		ID:        chat.ID,
		Title:     chat.Title,
		CreatedAt: chat.CreatedAt,
		UpdatedAt: chat.UpdatedAt,
		Couple:    make([]result.Couple, 0, len(chat.Couples)),
	}

	for _, c := range chat.Couples {
		r := result.Couple{
			Command: result.Command{
				ID:        c.Command.ID,
				Text:      c.Command.Prompt,
				CreatedAt: c.Command.CreatedAt,
				Status:    c.Command.Status,
				Media:     make([]string, 0, len(c.CommandMedia)),
			},
			Result: result.Result{
				ID:        c.Result.ID,
				Text:      c.Result.Text,
				CreatedAt: c.Result.CreatedAt,
				Media:     make([]string, 0, len(c.ResultMedia)),
			},
		}

		for _, cm := range c.CommandMedia {
			r.Command.Media = append(r.Command.Media, cm.URL.String())
		}

		for _, cm := range c.ResultMedia {
			r.Result.Media = append(r.Result.Media, cm.URL.String())
		}

		res.Couple = append(res.Couple, r)
	}

	return []result.GetChats{res}, nil
}

func (c *Chatting) UpdateTitle(ctx context.Context, cm command.UpdateTitle) error {
	const op = "application.service.Chatting.Update"

	chat, err := c.repo.GetChatByID(ctx, cm.ChatID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	chat.Title = cm.Title

	err = c.repo.SaveChat(ctx, chat)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (c *Chatting) DeleteChat(ctx context.Context, cm command.DeleteChat) error {
	const op = "application.service.Chatting.Delete"

	err := c.repo.DeleteChatByID(ctx, cm.ChatID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (c *Chatting) ExecuteCommand(ctx context.Context, cm command.Execute) (*event.Manager, error) {
	const op = "application.service.Chatting.Execute"

	var plan *entities.Plan
	var userID *uuid.UUID

	userID, ok := models.GetUserIDFromCtx(ctx)
	if ok {
		u, err := c.authRepo.GetUserByID(ctx, *userID)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}

		plan = u.Plan
	}

	if ip, ok := models.GetUnloggedUserIPFromCtx(ctx); ok {
		u, err := c.authRepo.GetUnloggedByIP(ctx, *ip)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}

		plan = u.Plan
	}

	if plan == nil {
		return nil, errors.New("unlogged user not found")
	}

	return c.execute(ctx, cm, plan, userID)
}
