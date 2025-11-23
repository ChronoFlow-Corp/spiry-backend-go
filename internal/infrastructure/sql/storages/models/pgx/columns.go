package pgx

type columnIndex uint

const (
	id columnIndex = iota
	name
	minLevel
	createdAt
	updatedAt
)

var table = "models"

var columns = []string{
	id:        "id",
	name:      "name",
	minLevel:  "min_level",
	createdAt: "created_at",
	updatedAt: "updated_at",
}
