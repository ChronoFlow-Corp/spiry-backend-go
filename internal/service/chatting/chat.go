package chatting

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/repository/commands"
	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/repository/entities"
	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/service"
	"github.com/ChronoFlow-Corp/spiry-backend-go/pkg/llm"
	"github.com/ChronoFlow-Corp/spiry-backend-go/pkg/slctx"
	"github.com/google/uuid"
)

type chatProvider interface {
	CreateChat(ctx context.Context, chat entities.Chat) error
	GetChatByID(ctx context.Context, chatID uuid.UUID) (entities.Chat, error)
	UpdateChat(ctx context.Context, id uuid.UUID, title string) error
	DeleteChat(ctx context.Context, id uuid.UUID) error
	GetChatByUserID(ctx context.Context) ([]entities.Chat, error)
}

type messageProvider interface {
	AddMessage(ctx context.Context, msg entities.Message) error
}

type userProvider interface {
	GetUserByID(ctx context.Context) (entities.User, error)
	SetPlanLimit(ctx context.Context, limit int, planID uuid.UUID) error
}

type toolProvider interface {
	GetToolByName(ctx context.Context, name string) (entities.Tool, error)
}

type txProvider interface {
	WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}

type Service struct {
	cl *llm.Client
	ch chatProvider
	mp messageProvider
	tl toolProvider
	tx txProvider
	up userProvider
}

func NewChat(
	cl *llm.Client,
	ch chatProvider,
	mp messageProvider,
	tl toolProvider,
	tx txProvider,
	up userProvider,
) Service {
	return Service{cl: cl, ch: ch, mp: mp, tl: tl, tx: tx, up: up}
}

func (s Service) GetChats(ctx context.Context, id *uuid.UUID) ([]entities.Chat, error) {
	const op = "service.chatting.GetChats"

	if id == nil {
		chats, err := s.ch.GetChatByUserID(ctx)
		if err != nil {
			slctx.Logger(ctx).Debug("cannot get chats", slog.Any("err", err))
			return nil, fmt.Errorf("%s: %w", op, err)
		}

		return chats, nil
	}

	chats := make([]entities.Chat, 0)

	chat, err := s.ch.GetChatByID(ctx, *id)
	if err != nil {
		slctx.Logger(ctx).Debug("cannot get chat", slog.Any("err", err))
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	chats = append(chats, chat)

	return chats, nil
}

func (s Service) UpdateChat(ctx context.Context, chat commands.UpdateChatTitle) error {
	const op = "service.chatting.UpdateChat"

	s.ch.UpdateChat(ctx, chat.ChatID, chat.NewTitle)
	return nil
}

func (s Service) DeleteChat(ctx context.Context, id uuid.UUID) error {
	const op = "service.chatting.DeleteChat"

	err := s.ch.DeleteChat(ctx, id)
	if err != nil {
		slctx.Logger(ctx).Debug("cannot delete chat", slog.Any("err", err))

		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func generatePrompt(t entities.Tool, vars map[string]string) llm.Prompt {
	for k, v := range t.Vars {
		t.Prompt = strings.ReplaceAll(t.Prompt, v, vars[k])
	}

	return llm.Prompt{
		Model: t.Model,
		Messages: []llm.Message{
			{
				Role:    "user",
				Content: t.Prompt,
			},
		},
	}
}

func (s Service) checkLimits(ctx context.Context) error {
	const op = "service.chatting.checkLimits"

	u, err := s.up.GetUserByID(ctx)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	switch u.Plan.Name {
	case entities.FreePlanName:
		if u.Plan.Limit != nil && *u.Plan.Limit != 0 {
			return fmt.Errorf("%s: %w", op, service.LimitExceededErr)
		}

		s.up.SetPlanLimit(ctx, *u.Plan.Limit-1, u.Plan.ID)

		return nil
	case entities.ProPlanName:
		if u.Plan.End != nil && u.Plan.End.Before(time.Now()) {
			return fmt.Errorf("%s: %w", op, service.ProPlanExpiredErr)
		}
	}

	return nil
}
