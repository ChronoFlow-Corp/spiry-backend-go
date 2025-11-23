package pgx

type columnIndex uint

const (
	id columnIndex = iota
	name
	modalities
	settings
	prompt
	minLevel
	createdAt
	updatedAt
)

const table = "tools"

var columns = []string{
	id:         "id",
	name:       "name",
	modalities: "modalities",
	settings:   "settings",
	prompt:     "prompt",
	minLevel:   "min_level",
	createdAt:  "created_at",
	updatedAt:  "updated_at",
}
