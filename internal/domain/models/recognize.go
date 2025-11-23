package models

import "github.com/ChronoFlow-Corp/spiry-backend-go/internal/domain/entities"

type RecognizeResult struct {
	Model *entities.Model
	Tool  *entities.Tool
	Title string
}

type RecognizeCommand struct {
	Prompt        string
	AllowedTools  []*entities.Tool
	AllowedModels []*entities.Model
}
