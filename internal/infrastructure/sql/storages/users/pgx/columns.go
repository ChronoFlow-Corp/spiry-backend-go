package pgx

type columnIndex = uint

const (
	id columnIndex = iota
	email
	avatarURL
	name
	lastName
	theme
	googleAccessToken
	googleRefreshToken
	createdAt
	updatedAt
)

var table = "users"

var columns = []string{
	id:                 "id",
	email:              "email",
	avatarURL:          "avatar_url",
	name:               "name",
	lastName:           "last_name",
	theme:              "theme",
	googleAccessToken:  "google_access_token",
	googleRefreshToken: "google_refresh_token",
	createdAt:          "created_at",
	updatedAt:          "updated_at",
}
