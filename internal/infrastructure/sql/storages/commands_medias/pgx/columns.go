package pgx

type columnIndex uint

const (
	id columnIndex = iota
	name
	mediaType
	indexUrl
	size
	commandID
	userID
	createdAt
	updatedAt
)

var table = "command_medias"

var columns = []string{
	id:        "id",
	name:      "name",
	mediaType: "type",
	indexUrl:  "url",
	size:      "size",
	commandID: "command_id",
	userID:    "user_id",
	createdAt: "created_id",
	updatedAt: "updated_id",
}
