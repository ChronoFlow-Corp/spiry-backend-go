package pgx

type columnIndex = uint

const (
	id columnIndex = iota
	quote
	userID
	subscriptionID
	createdAt
	updatedAt
)

var table = "plans"

var columns = []string{
	id:             "id",
	quote:          "quote",
	userID:         "user_id",
	subscriptionID: "subscription_id",
	createdAt:      "created_at",
	updatedAt:      "updated_at",
}
