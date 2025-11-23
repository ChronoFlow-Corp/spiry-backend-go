package chatting

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/application/command"
	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/application/pkg/event"
	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/domain/aggregates"
	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/domain/entities"
	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/domain/models"
	"github.com/ChronoFlow-Corp/spiry-backend-go/pkg/slctx"
	"github.com/google/uuid"
)

func (c *Chatting) execute(ctx context.Context, cm command.Execute) (*event.Manager, error) {
	const op = "service.chatting.execute"

	var chat *aggregates.Chat
	var user *aggregates.User
	var err error

	userID, ok := models.GetUserIDFromCtx(ctx)
	if ok {
		user, err = c.authRepo.GetUserByID(ctx, *userID)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
	}

	if cm.ChatID != nil {
		chat, err = c.repo.GetChatByID(ctx, *cm.ChatID)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
	} else {
		chat, err = aggregates.NewChat(entities.NewChat(userID, "New chat"), nil)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
	}

	aggCommand, err := c.newCommandFromInput(ctx, cm, chat.ID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	allowedTools, err := c.repo.GetAllowedTools(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	allowedModels, err := c.repo.GetAllowedModels(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	recognized, err := c.chattingDomain.RecognizeTool(ctx, models.RecognizeCommand{
		Prompt:        aggCommand.Prompt,
		AllowedTools:  allowedTools,
		AllowedModels: allowedModels,
	})
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	err = c.repo.SaveChat(ctx, chat)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	err = c.repo.SaveCommand(ctx, aggCommand)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	stream, err := c.chattingDomain.ExecuteStreaming(ctx, models.ExecuteStreaming{
		Command: aggCommand,
		Model:   recognized.Model,
		Tool:    recognized.Tool,
		Context: chat.Couples,
	})
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	eventer := event.NewManager(ctx, 10)

	go func() {
		err := c.saveResult(aggCommand, recognized, user, chat, stream, eventer)
		if err != nil {
			slctx.Logger(ctx).Debug("cannot save result", slog.Any("error", err))
		}
	}()

	return eventer, nil
}

func (c *Chatting) newCommandFromInput(
	ctx context.Context,
	cm command.Execute,
	chatID uuid.UUID,
) (*aggregates.Command, error) {

	userID, _ := models.GetUserIDFromCtx(ctx)

	medias, err := c.repo.GetMediaByUrls(ctx, cm.Media)
	if err != nil {
		return nil, err
	}

	var aggCommand *aggregates.Command
	switch {
	case cm.ToolName == "" && cm.ModelName == "":
		cmEntity := entities.NewCommand(
			cm.Prompt,
			cm.Settings,
			cm.Flags,
			entities.StatusInProgress,
			nil,
			chatID,
			nil,
			userID,
		)

		aggCommand, err = aggregates.NewCommand(cmEntity, medias, nil, nil)
		if err != nil {
			return nil, err
		}
	case cm.ToolName == "" && cm.ModelName != "":
		model, err := c.repo.GetModelByName(ctx, cm.ModelName)
		if err != nil {
			return nil, err
		}

		cmEntity := entities.NewCommand(
			cm.Prompt,
			cm.Settings,
			cm.Flags,
			entities.StatusInProgress,
			&model.ID,
			chatID,
			nil,
			userID,
		)

		aggCommand, err = aggregates.NewCommand(cmEntity, medias, model, nil)
		if err != nil {
			return nil, err
		}
	case cm.ToolName != "" && cm.ModelName == "":
		tool, err := c.repo.GetToolByName(ctx, cm.ToolName)
		if err != nil {
			return nil, err
		}

		cmEntity := entities.NewCommand(
			cm.Prompt,
			cm.Settings,
			cm.Flags,
			entities.StatusInProgress,
			nil,
			chatID,
			&tool.ID,
			userID,
		)

		aggCommand, err = aggregates.NewCommand(cmEntity, medias, nil, tool)
		if err != nil {
			return nil, err
		}
	default:
		tool, err := c.repo.GetToolByName(ctx, cm.ToolName)
		if err != nil {
			return nil, err
		}

		model, err := c.repo.GetModelByName(ctx, cm.ModelName)
		if err != nil {
			return nil, err
		}

		cmEntity := entities.NewCommand(
			cm.Prompt,
			cm.Settings,
			cm.Flags,
			entities.StatusInProgress,
			&model.ID,
			chatID,
			&tool.ID,
			userID,
		)

		aggCommand, err = aggregates.NewCommand(cmEntity, medias, model, tool)
		if err != nil {
			return nil, err
		}
	}

	return aggCommand, nil
}
