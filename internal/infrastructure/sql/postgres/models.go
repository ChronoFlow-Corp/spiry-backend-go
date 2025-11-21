package postgres

import (
	"context"
	"fmt"

	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/domain/entities"
	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/infrastructure/sql/models"
	"github.com/jackc/pgx/v5"
)

func (d *Database) GetModels(ctx context.Context) ([]*entities.Model, error) {
	const op = "infrastructure.sql.postgres.GetModels"

	q := `select 
id as model_id, 
name as model_name, 
modalities as model_modalities, 
min_level as model_min_level, 
created_at as model_created_at, 
updated_at as model_updated_at
from models`

	rows, err := d.pool.Query(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()
	dbModels, err := pgx.CollectRows(rows, pgx.RowToStructByName[models.Model])
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	modelsEntities := make([]*entities.Model, 0, len(dbModels))
	for _, m := range dbModels {
		modalities := make([]entities.Modality, 0, len(m.Modalities))
		for _, m := range m.Modalities {
			modalities = append(modalities, entities.Modality(m))
		}
		modelEntity := entities.NewModel(m.Name, modalities, m.MinLevel)
		modelsEntities = append(modelsEntities, modelEntity)
	}

	return modelsEntities, nil
}

func (d *Database) GetModelByName(ctx context.Context, name string) (*entities.Model, error) {
	const op = "infrastructure.sql.postgres.GetModelByName"

	q := `select 
id as model_id, 
name as model_name, 
modalities as model_modalities, 
min_level as model_min_level, 
created_at as model_created_at, 
updated_at as model_updated_at
from models where name = $1`

	rows, err := d.pool.Query(ctx, q, name)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()
	dbModels, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[models.Model])
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	mod := make([]entities.Modality, 0, len(dbModels.Modalities))
	for _, m := range dbModels.Modalities {
		mod = append(mod, entities.Modality(m))
	}

	modelEnt := &entities.Model{
		ID:         dbModels.ID,
		Name:       dbModels.Name,
		Modalities: mod,
		MinLevel:   dbModels.MinLevel,
		CreatedAt:  dbModels.CreatedAt,
		UpdatedAt:  dbModels.UpdatedAt,
	}
	return modelEnt, nil
}
