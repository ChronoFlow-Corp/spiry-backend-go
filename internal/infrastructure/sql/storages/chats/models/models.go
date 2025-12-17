package models

import (
	"time"

	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/infrastructure/sql/storages/commands/models"
	commandMedia "github.com/ChronoFlow-Corp/spiry-backend-go/internal/infrastructure/sql/storages/commands_medias/models"
	stmodel "github.com/ChronoFlow-Corp/spiry-backend-go/internal/infrastructure/sql/storages/models/models"
	resultMedia "github.com/ChronoFlow-Corp/spiry-backend-go/internal/infrastructure/sql/storages/result_medias/models"
	result "github.com/ChronoFlow-Corp/spiry-backend-go/internal/infrastructure/sql/storages/results/models"
	"github.com/google/uuid"
)

type Chat struct {
	ID        uuid.UUID  `db:"id"`
	Title     string     `db:"title"`
	CreatedAt time.Time  `db:"created_at"`
	UpdatedAt time.Time  `db:"updated_at"`
	UserID    *uuid.UUID `db:"user_id"`
}

type ChatWithCouples struct {
	Chat

	CommandResultCouple []byte `db:"command_result_pairs"`
}

type Couple struct {
	Command CommandWithMedias `json:"command"`
	Result  ResultWithMedias  `json:"result"`
}

type CommandWithMedias struct {
	models.Command

	Medias []commandMedia.CommandMedia `json:"medias"`
}

type ResultWithMedias struct {
	result.Result

	Medias []resultMedia.ResultMedia `json:"medias"`
	Model  stmodel.Model             `json:"model"`
}
