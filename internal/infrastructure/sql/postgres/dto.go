package postgres

import "github.com/ChronoFlow-Corp/spiry-backend-go/internal/infrastructure/sql/models"

type userDTO struct {
	models.User
	models.Session
	models.Plan
	models.Model
}

type chatDto struct {
	models.Chat
	Couples []couplesDto `db:"command_result_pairs"`
}

type couplesDto struct {
	Command commandWithMedia `json:"command"`
	Result  resultWithMedia  `json:"result"`
}

type commandWithMedia struct {
	models.Command
	Medias []models.CommandMedia `json:"medias"`
}

type resultWithMedia struct {
	models.Result
	Medias []models.ResultMedia `json:"medias"`
}

const toolRows = `
tools.id as tool_id, 
tools.name as tool_name, 
tools.modalities as tool_modalities,
tools.settings as tool_settings,
tools.prompt as tool_prompt,
tools.min_level as tool_min_level,
tools.created_at as tool_created_at,
tools.updated_at as tool_updated_at 
`
