package chatting

import (
	"context"
	"errors"
	"fmt"

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
	ctx context.Context,
	commandAggregate *aggregates.Command,
	rec models2.RecognizeResult,
	userID *uuid.UUID,
	chat *aggregates.Chat,
	stream *pubSub.PubSub[entities.Chunk],
	eventer *event.Manager,
) (err error) {
	const op = "application.service.Chatting.saveResult"

	resultID := uuid.New()
	collector := result_collector.NewCollector(userID, resultID, chat.ID)
	st := make(chan entities.Chunk)

	defer eventer.Close()
	defer func() {
		if err != nil {
			sendEvent(eventer, model.NewErrorEvent(chat.ID, resultID, err))
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

		sendEvent(eventer, model.NewGeneratingEvent(chat.ID, resultID, chunk.Content))
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

		sendEvent(eventer, model.NewTitleEvent(chat.ID, resultID, rec.Title))
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

	sendEvent(eventer, model.NewDoneEvent(chat.ID, resultID))

	return nil
}

func sendEvent(eventer *event.Manager, event model.Event) {
	if !eventer.Closed() {
		err := eventer.Send(event)
		if err != nil {
			return
		}
	}
}
