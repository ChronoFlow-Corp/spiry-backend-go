package chatting

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/repository/commands"
	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/repository/entities"
	"github.com/ChronoFlow-Corp/spiry-backend-go/pkg/llm"
	"github.com/ChronoFlow-Corp/spiry-backend-go/pkg/slctx"
	"github.com/google/uuid"
)

func (s Service) SendMessage(ctx context.Context, cm commands.SendMessage) (*Socket, error) {
	const op = "service.Service"


	userID, ok := ctx.Value(entities.UserIDCtxKey{}).(uuid.UUID)
	if !ok {
		return nil, fmt.Errorf("%s: %w", op, errors.New("missing user ID in context"))
	}

	tool, err := s.tl.GetToolByName(ctx, cm.Tool)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", err, op)
	}

	if cm.ChatID != nil {
		err := s.checkLimits(ctx)
		if err != nil {
			return nil, fmt.Errorf("%w: %s", err, op)
		}

		chat, err := s.ch.GetChatByID(ctx, *cm.ChatID)
		if err != nil {
			return nil, fmt.Errorf("%w: %s", err, op)
		}

		messageID := uuid.New()

		saveCtx := context.WithValue(context.Background(), entities.UserIDCtxKey{}, userID)
		saveCtx, cancel := context.WithTimeout(saveCtx, time.Minute)

		ch, err := s.cl.DoStream(ctx, generatePrompt(tool, cm.Vars))
		if err != nil {
			cancel()

			return nil, fmt.Errorf("%w: %s", err, op)
		}

		soc := NewSocket(chat.ID, messageID, ch)

		go s.saveMessageLogged(ctx, cm, soc, messageID, ch, cancel)

		return soc, nil
	}

	err = s.checkLimits(ctx)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", err, op)
	}

	chat := entities.NewChat(uuid.New(), "New chat", tool.ID, userID, nil)

	err = s.ch.CreateChat(ctx, chat)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", err, op)
	}

	saveCtx := context.WithValue(context.Background(), entities.UserIDCtxKey{}, userID)
	saveCtx, cancel := context.WithTimeout(saveCtx, time.Minute)

	ch, err := s.cl.DoStream(saveCtx, generatePrompt(tool, cm.Vars))
	if err != nil {
		cancel()

		return nil, fmt.Errorf("%w: %s", err, op)
	}

	messageID := uuid.New()
	
	soc := NewSocket(chat.ID, messageID, ch)

	go func() {
		ch.Done()
		soc.sendWarning(WarningEvent{
			Err: nil,
			Cause: DoneCause,
			Description: "Done generating message",
		})
	}()



	go s.saveMessageLogged(saveCtx, cm, soc, messageID, ch, cancel)

	return soc, nil
}

func (s Service) saveMessageLogged(
	ctx context.Context,
	cm commands.SendMessage,
	soc *Socket,
	msgID uuid.UUID,
	l llm.ChunkReader,
	cancelFn context.CancelFunc) {
	const op = "service.saveMessageLogged"

	defer cancelFn()

	chunks := make([]llm.Chunk, 0)
	for chunk, ok := l.Read(); ok; chunk, ok = l.Read() {
		chunks = append(chunks, chunk)
		soc.sendInfo(InfoEvent{
			ID:      msgID,
			ChatID: soc.chatID,
			Content: chunk.Content.Content,
			Role: chunk.Content.Role,
		})
	}

	slctx.Logger(ctx).Debug("finished reading chunks", slog.Int("count", len(chunks)))

	var content string

	var role string

	for _, chunk := range chunks {
		content += chunk.Content.Content
		role = chunk.Content.Role
	}


	err := s.mp.AddMessage(ctx, entities.Message{
		ID:        uuid.New(),
		ChatID:    soc.chatID,
		Text: cm.Vars["content"],
		UserID:    ctx.Value(entities.UserIDCtxKey{}).(uuid.UUID),
		Role:      cm.Role,
	})
	if err != nil {
		slctx.Logger(ctx).Debug("cannot save user message", slog.Any("error", err))

		soc.sendError(fmt.Errorf("%w: %s", err, op))

		return
	}
	
	err = s.mp.AddMessage(ctx, entities.Message{
		ID:        msgID,
		Text: content,
		Role:    role,
		ChatID: soc.chatID,
		UserID: ctx.Value(entities.UserIDCtxKey{}).(uuid.UUID),
	})
	if err != nil {
		slctx.Logger(ctx).Debug("cannot save model message", slog.Any("error", err))

		soc.sendError(fmt.Errorf("%w: %s", err, op))

		return
	}
}

func (s Service) sendMessageUnlogged(ctx context.Context, cm commands.SendMessage) {
	const op = "service.SendMessageUnlogged"
}
