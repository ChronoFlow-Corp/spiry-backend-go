package pgx

type columnIndex = uint

const (
	id columnIndex = iota
	openRouterID
	text
	commandID
	toolID
	chatID
	userID
	modelID
	createdAt
	updatedAt
)

var table = "results"

var columns = []string{
	id:           "id",
	openRouterID: "open_router_id",
	text:         "text",
	commandID:    "command_id",
	toolID:       "tool_id",
	chatID:       "chat_id",
	userID:       "user_id",
	modelID:      "model_id",
	createdAt:    "created_at",
	updatedAt:    "updated_at",
}
