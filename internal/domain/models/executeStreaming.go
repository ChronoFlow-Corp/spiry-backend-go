package models

import (
	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/domain/aggregates"
	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/domain/entities"
)

type ExecuteStreaming struct {
	Command *aggregates.Command
	Model   *entities.Model
	Tool    *entities.Tool
	Context []aggregates.CommandResultCouple
}
