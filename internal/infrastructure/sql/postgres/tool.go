package postgres

import (
	"context"
	"fmt"

	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/domain/aggregates"
	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/domain/entities"
	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/infrastructure/sql/models"
	"github.com/jackc/pgx/v5"
)

func (d *Database) GetToolByName(ctx context.Context, name string) (*entities.Tool, error) {
	const op = "infrastructure.sql.postgres.GetToolByName"

	q := `select 
id as tool_id, 
name as tool_name, 
modalities as tool_modalities,
settings as tool_settings,
prompt as tool_prompt,
min_level as tool_min_level,
created_at as tool_created_at,
updated_at as tool_updated_at 
from tools where name = $1`

	row, err := d.pool.Query(ctx, q, name)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer row.Close()

	t, err := pgx.CollectOneRow(row, pgx.RowToStructByName[models.Tool])
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	mod := make([]entities.Modality, 0, len(t.Modalities))
	for _, modality := range t.Modalities {
		mod = append(mod, entities.Modality(modality))
	}

	tool := &entities.Tool{
		ID:         t.ID,
		Name:       t.Name,
		Modalities: mod,
		Settings:   t.Settings,
		Prompt:     t.Prompt,
		MinLevel:   uint(t.MinLevel),
		CreatedAt:  t.CreatedAt,
		UpdatedAt:  t.UpdatedAt,
	}

	return tool, nil
}

func (d *Database) GetAllowedTools(ctx context.Context, us *aggregates.User) ([]*entities.Tool, error) {
	const op = "infrastructure.sql.postgres.GetAllowedTools"

	q := fmt.Sprintf(`
select %s from tools 
left join plans on plans.user_id = $1 
left join subscriptions on plans.subscription_id = subscriptions.id 
where subscriptions.level >= tools.min_level
`, toolRows)

	rows, err := d.pool.Query(ctx, q, us.ID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	dtos, err := pgx.CollectRows(rows, pgx.RowToStructByName[models.Tool])
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	tools := make([]*entities.Tool, 0, len(dtos))

	for _, dto := range dtos {
		mod := make([]entities.Modality, 0, len(dto.Modalities))
		for _, modality := range dto.Modalities {
			mod = append(mod, entities.Modality(modality))
		}

		tool := &entities.Tool{
			ID:         dto.ID,
			Name:       dto.Name,
			Modalities: mod,
			Settings:   dto.Settings,
			Prompt:     dto.Prompt,
			MinLevel:   uint(dto.MinLevel),
			CreatedAt:  dto.CreatedAt,
			UpdatedAt:  dto.UpdatedAt,
		}

		tools = append(tools, tool)
	}

	return tools, nil
}

func (d *Database) GetTools(ctx context.Context) ([]*entities.Tool, error) {
	const op = "infrastructure.sql.postgres.GetTools"

	q := `select
id as tool_id,
name as tool_name,
modalities as tool_modalities,
settings as tool_settings,
prompt as tool_prompt,
min_level as tool_min_level,
created_at as tool_created_at,
updated_at as tool_updated_at
from tools`

	rows, err := d.pool.Query(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	ts, err := pgx.CollectRows(rows, pgx.RowToStructByName[models.Tool])
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	tools := make([]*entities.Tool, 0, len(ts))
	for _, t := range ts {
		mod := make([]entities.Modality, 0, len(t.Modalities))
		for _, modality := range t.Modalities {
			mod = append(mod, entities.Modality(modality))
		}

		tool := &entities.Tool{
			ID:         t.ID,
			Name:       t.Name,
			Modalities: mod,
			Settings:   t.Settings,
			Prompt:     t.Prompt,
			MinLevel:   uint(t.MinLevel),
			CreatedAt:  t.CreatedAt,
			UpdatedAt:  t.UpdatedAt,
		}

		tools = append(tools, tool)
	}

	return tools, nil
}
