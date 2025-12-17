package aggregates

import (
	"errors"
	"fmt"

	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/domain/entities"
)

type Command struct {
	*entities.Command

	Medias []*entities.CommandMedia
	Tool   *entities.Tool

	// The model can be nil if user did not specify specific model.
	// If nil, the model choose automatically by the system.
	Model *entities.Model
}

func NewCommand(
	command *entities.Command,
	medias []*entities.CommandMedia,
	model *entities.Model,
	tool *entities.Tool,
) (*Command, error) {
	const op = "aggregates.Command.NewCommand"

	err := command.Validate()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	for _, media := range medias {
		err := media.Validate()
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
	}

	if tool != nil {
		err = tool.Validate()
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
	}

	if model != nil {
		err = model.Validate()
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
	}

	return &Command{
		Command: command,
		Medias:  medias,
		Tool:    tool,
		Model:   model,
	}, nil
}

func (c *Command) SetModel(model *entities.Model) error {
	const op = "aggregates.Command.SetModel"

	if model == nil {
		return fmt.Errorf("%s: %w", op, errors.New("model is nil"))
	}

	err := model.Validate()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (c *Command) SetStatus(status string) error {
	const op = "aggregates.Command.SetStatus"

	c.Status = status

	err := c.Command.Validate()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
