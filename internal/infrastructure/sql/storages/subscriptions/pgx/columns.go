package pgx

type columnIndex uint

const (
	id columnIndex = iota
	name
	modalitiesQuote
	period
	price
	level
	createdAt
	updatedAt
)

var table = "subscriptions"

var columns = []string{
	id:              "id",
	name:            "name",
	modalitiesQuote: "modalities_quote",
	period:          "period",
	price:           "price",
	level:           "level",
	createdAt:       "created_at",
	updatedAt:       "updated_at",
}
