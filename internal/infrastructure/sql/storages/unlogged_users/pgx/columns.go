package pgx

type columnIndex = uint

const (
	id columnIndex = iota
	ip_address
	planID
	createdAt
)

var table = "unlogged_users"

var columns = []string{
	id:         "id",
	ip_address: "ip_address",
	planID:     "plan_id",
	createdAt:  "created_at",
}
