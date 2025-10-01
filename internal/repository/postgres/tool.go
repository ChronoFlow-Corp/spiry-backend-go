package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/repository"
	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/repository/entities"
)

func (p *Postgres) GetToolByName(ctx context.Context, name string) (entities.Tool, error) {
	const op = "repository.postgres.GetToolByName"

	query := `SELECT id, name, model, prompt, vars, created_at, updated_at
			  FROM tools
			  WHERE name = $1`

	var tool dbTool

	err := p.db.GetContext(ctx, &tool, query, name)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entities.Tool{}, repository.NewNotFound(err, "tool not found", "tools.name", name)
		}
		return entities.Tool{}, fmt.Errorf("%s: %w", op, err)
	}

	return dbToolToRepositoryTool(tool), nil
}