package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/repository"
	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/service/policy"
	"github.com/ChronoFlow-Corp/spiry-backend-go/pkg/llm"
	"github.com/google/uuid"
	"time"
)

type chatProvider interface {
	CreateChat(ctx context.Context, chat repository.Chat) error
}

type messageProvider interface {
	AddMessage(ctx context.Context, chatId uuid.UUID, msg repository.Message) error
}

type Chat struct {
	cl *llm.Client
	cr chatProvider
	mp messageProvider
	p  policy.Prompt
}

func NewChat(cl *llm.Client, p policy.Prompt, ch chatProvider, mp messageProvider) Chat {
	return Chat{cl: cl, p: p, cr: ch, mp: mp}
}

func (c Chat) DefaultChat(ctx context.Context, cm repository.ChatCommand) (llm.ChunkReaderCloser, *WsError) {
	const op = "service.Chat"
	vars := cm.GetVars()
	wsErr := newWsError(nil)
	pr, err := c.p.Get(cm.Type, vars)
	if err != nil {
		wsErr.SetError(fmt.Errorf("%s: %w", op, err))
		return nil, wsErr
	}

	res, err := c.cl.DoStream(ctx, pr)
	if err != nil {
		wsErr.SetError(fmt.Errorf("%s: %w", op, err))
		return nil, wsErr
	}

	if cm.UserID != nil {
		go func() {
			res.Wait()
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
			defer cancel()

			msg := res.Summary()

			chatID, err := uuid.Parse(cm.ChatID)
			if err != nil {
				wsErr.SetError(fmt.Errorf("%s: %w", op, err))

				return
			}

			err = c.cr.CreateChat(ctx, repository.NewChat(chatID, *cm.UserID, msg.Title))

			var uniqErr *repository.ErrorUnique

			if err != nil && errors.As(err, &uniqErr) {
				err = c.mp.AddMessage(ctx, chatID, repository.NewMessage(uuid.New(), cm.Text, msg.CompleteMessage, *cm.UserID))
				if err != nil {
					wsErr.SetError(fmt.Errorf("%s: %w", op, err))

					return
				}
			}

			if err != nil {
				wsErr.SetError(fmt.Errorf("%s: %w", op, err))

				return
			}
		}()
	}

	//TODO: create chat with ai generated title
	return res, wsErr
}
