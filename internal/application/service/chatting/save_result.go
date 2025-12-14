package chatting

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/application/model"
	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/application/pkg/event"
	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/application/pkg/pubSub"
	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/application/pkg/result_collector"
	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/domain"
	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/domain/aggregates"
	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/domain/entities"
	models2 "github.com/ChronoFlow-Corp/spiry-backend-go/internal/domain/models"
	"github.com/google/uuid"
)

func (c *Chatting) saveResult(
	commandAggregate *aggregates.Command,
	rec models2.RecognizeResult,
	userID *uuid.UUID,
	chat *aggregates.Chat,
	stream *pubSub.PubSub[entities.Chunk],
	eventer *event.Manager,
) (err error) {
	const op = "application.service.Chatting.saveResult"

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	resultID := uuid.New()
	collector := result_collector.NewCollector(userID, resultID, chat.ID)
	st := make(chan entities.Chunk)

	defer eventer.Close()
	defer func() {
		if err != nil {
			_ = eventer.Send(model.NewErrorEvent(chat.ID, resultID, err)) //nolint:errcheck
		}
	}()

	stream.Subscribe(st)

	for chunk := range st {
		if chunk.Err != nil {
			if errors.Is(chunk.Err, domain.ErrStreamClosed) {
				break
			}

			return fmt.Errorf("%s: %w", op, chunk.Err)
		}

		collector.AddChunk(chunk)

		if !eventer.Closed() {
			_ = eventer.Send( //nolint:errcheck
				model.NewGeneratingEvent(chat.ID, resultID, chunk.Content),
			)
		}
	}

	if len(chat.Couples) == 0 {
		err = chat.SetTitle(rec.Title)
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}

		err = c.repo.SaveChat(ctx, chat)
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}

		if !eventer.Closed() {
			_ = eventer.Send(model.NewTitleEvent(chat.ID, resultID, rec.Title)) //nolint:errcheck
		}
	}

	openRouterID, fullAnswer := collector.Finalize()

	resEntity := entities.NewResult(
		fullAnswer,
		openRouterID,
		commandAggregate.ID,
		rec.Tool.ID,
		chat.ID,
		rec.Model.ID,
		userID,
	)

	res, err := aggregates.NewResult(resEntity, nil, rec.Tool, rec.Model)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	err = c.repo.SaveResult(ctx, res)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	commandAggregate.Status = entities.StatusDone

	err = c.repo.SaveCommand(ctx, commandAggregate)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if !eventer.Closed() {
		_ = eventer.Send(model.NewDoneEvent(chat.ID, resultID)) //nolint:errcheck
	}

	return nil
}
