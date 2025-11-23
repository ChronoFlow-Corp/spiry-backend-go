package aggregates

import "github.com/ChronoFlow-Corp/spiry-backend-go/internal/domain/entities"

type Result struct {
	*entities.Result

	Medias []*entities.ResultMedia
	Tool   *entities.Tool
	Model  *entities.Model
}

func NewResult(
	result *entities.Result,
	medias []*entities.ResultMedia,
	tool *entities.Tool,
	model *entities.Model) (*Result, error) {
	err := result.Validate()
	if err != nil {
		return nil, err
	}

	for _, media := range medias {
		err = media.Validate()
		if err != nil {
			return nil, err
		}
	}

	err = tool.Validate()
	if err != nil {
		return nil, err
	}

	err = model.Validate()
	if err != nil {
		return nil, err
	}

	return &Result{Result: result, Medias: medias, Tool: tool, Model: model}, nil
}
