package pgx

type columnIndex uint

const (
	id columnIndex = iota
	prompt
	settings
	flags
	chatID
	toolID
	modelID
	status
	userID
	createdAt
	updatedAt
)

var table = "commands"

var columns = []string{
	id:        "id",
	prompt:    "prompt",
	settings:  "settings",
	flags:     "flags",
	chatID:    "chat_id",
	toolID:    "tool_id",
	modelID:   "model_id",
	status:    "status",
	userID:    "user_id",
	createdAt: "created_at",
	updatedAt: "updated_at",
}
