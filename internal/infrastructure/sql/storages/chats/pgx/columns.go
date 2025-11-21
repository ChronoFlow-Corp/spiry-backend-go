package pgx

type columnIndex uint

const (
	id columnIndex = iota
	title
	createdAt
	updatedAt
	userID
)

const table = "chats"

var columns = []string{
	id:        "id",
	title:     "title",
	createdAt: "created_at",
	updatedAt: "updated_at",
	userID:    "user_id",
}
