package pgx

type columnIndex = uint

const (
	id columnIndex = iota
	token
	expiresAt
	lastLogin
	device
	userID
	createdAt
	updatedAt
)

var table = "sessions"

var columns = []string{
	id:        "id",
	token:     "token",
	expiresAt: "expires_at",
	lastLogin: "last_login",
	device:    "device",
	userID:    "user_id",
	createdAt: "created_at",
	updatedAt: "updated_at",
}
