package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/application/pkg/pubSub"
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
		// TODO: error type
		return nil, errors.New("chatting: command is nil")
	}

	if streaming.Command.Status == "done" {
		// TODO: error type
		return nil, errors.New("chatting: command is done")
	}

	stream, err := c.llm.GenerateStreaming(ctx, streaming)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return stream, nil
}
