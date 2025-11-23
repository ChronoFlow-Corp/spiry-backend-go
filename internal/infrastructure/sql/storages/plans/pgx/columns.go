package pgx

type columnIndex = uint

const (
	id columnIndex = iota
	modalitiesQuote
	userID
	subscriptionID
	createdAt
	updatedAt
)

var table = "plans"

var columns = []string{
	id:              "id",
	modalitiesQuote: "modalities_quote",
	userID:          "user_id",
	subscriptionID:  "subscription_id",
	createdAt:       "created_at",
	updatedAt:       "updated_at",
}
