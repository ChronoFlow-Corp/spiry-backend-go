package pgx

type columnIndex uint

const (
	id columnIndex = iota
	name
	mediaType
	columnURL
	size
	resultID
	userID
	createdAt
	updatedAt
)

const table = "result_medias"

var columns = []string{
	id:        "id",
	name:      "name",
	mediaType: "mediaType",
	columnURL: "columnURL",
	size:      "size",
	resultID:  "resultID",
	userID:    "userID",
	createdAt: "createdAt",
	updatedAt: "updatedAt",
}
