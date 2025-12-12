package service

import (
	"context"
	"errors"
	"fmt"
	"slices"

	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/application/pkg/pubSub"
	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/domain"
	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/domain/aggregates"
	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/domain/entities"
	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/domain/models"
)

type LLmClient interface {
	GenerateStreaming(
		ctx context.Context, streaming models.ExecuteStreaming) (*pubSub.PubSub[entities.Chunk], error)

	// RecognizeTool choose tool from prompt and model
	// If model or tool specified of user, it had priority and return tool and model from command.
	RecognizeTool(
		ctx context.Context, command models.RecognizeCommand) (models.RecognizeResult, error)
}

type Chatting struct {
	llm LLmClient
}

func NewChatting(llm LLmClient) *Chatting {
	return &Chatting{llm: llm}
}

func (c *Chatting) RecognizeTool(
	ctx context.Context,
	cm models.RecognizeCommand,
) (models.RecognizeResult, error) {
	const op = "domain.service.Chatting.RecognizeTool"

	res, err := c.llm.RecognizeTool(ctx, cm)
	if err != nil {
		return models.RecognizeResult{}, fmt.Errorf("%s: %w", op, err)
	}

	return res, nil
}

func (c *Chatting) ExecuteStreaming(
	ctx context.Context,
	streaming models.ExecuteStreaming,
) (*pubSub.PubSub[entities.Chunk], error) {
	const op = "domain.service.Chatting.ExecuteStreaming"

	if streaming.Command == nil {
		return nil, errors.New("chatting: command is nil")
	}

	if streaming.Command.Status == "done" {
		return nil, errors.New("chatting: command is done")
	}

	stream, err := c.llm.GenerateStreaming(ctx, streaming)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return stream, nil
}

func (c *Chatting) CheckExecuteAndChangeLimits(
	cm aggregates.Command,
	plan *entities.Plan,
	tool entities.Tool,
) error {
	const op = "domain.service.Chatting.CanExecute"

	for i, tl := range plan.Quote.ToolLimits {
		if tl.ID == tool.ID {
			for k := range cm.Settings {
				if !tl.SettingsLimit[k] {
					return fmt.Errorf(
						"%s: %w",
						op,
						domain.NewForbidden(nil, "tool setting forbidden", k, cm.Settings[k]),
					)
				}
			}

			if tl.Usage == 0 {
				return fmt.Errorf(
					"%s: %w",
					op,
					domain.NewForbidden(nil, "tool usage quota exceeded", "usage", tool.Name),
				)
			}

			plan.Quote.ToolLimits[i].Usage--
		}
	}

	mediaCount := make(map[string]int)

	for _, m := range cm.Medias {
		mediaCount[m.Type]++
		for i, ml := range plan.Quote.MediaLimit {
			if m.Type == ml.Type {
				if ml.Size < m.Size {
					return errors.New("media size not allowed")
				}

				if slices.Contains(tool.Modalities, entities.ImageModality) {
					if ml.Generate == 0 {
						return fmt.Errorf("%s: %w", op, domain.NewForbidden(
							nil,
							"generation quota has been exceeded",
							"generate",
							ml.Type,
						))
					}

					plan.Quote.MediaLimit[i].Generate--
				}
			}
		}
	}

	for k, v := range mediaCount {
		for _, ml := range plan.Quote.MediaLimit {
			if k == ml.Type {
				if v > ml.Upload {
					return errors.New("count media upload not allowed")
				}
			}
		}
	}

	return nil
}
