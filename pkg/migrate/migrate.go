package migrate

import (
	"github.com/jmoiron/sqlx"
)

func Migrate(db *sqlx.DB, source string, steps int) error {
	return nil
}

