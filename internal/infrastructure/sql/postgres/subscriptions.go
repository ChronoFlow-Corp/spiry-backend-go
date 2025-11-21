package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/domain/entities"
	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/infrastructure/sql/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (d *Database) GetBaseSubscription(ctx context.Context) (*entities.Subscription, error) {
	const op = "infrastructure.sql.postgres.GetBaseSubscription"

	q := `select 
id as subscription_id,
name as subscription_name, 
modalities_quote as subscription_modalities_quote,
period as subscription_period, 
price as subscription_price,
level as subscription_level, 
created_at as subscription_created_at, 
updated_at as subscription_updated_at
from subscriptions where name = 'base'`

	rows, err := d.pool.Query(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	sub, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[models.Subscription])
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			if d.env == "development" {
				err := d.createDevBaseSubscription(ctx)
				if err != nil {
					return nil, fmt.Errorf("%s: %w", op, err)
				}

				return d.GetBaseSubscription(ctx)
			}
		}

		return nil, fmt.Errorf("%s: %w", op, err)
	}

	quoteEntity := entities.ModalitiesQuote{}

	if sub.ModalitiesQuote != nil && sub.ModalitiesQuote.ChattingQuote.Valid {
		var quote uint
		if sub.ModalitiesQuote.ChattingQuote.Int64 < 0 {
			quote = 0
			quoteEntity.ChattingQuote = &quote
		} else {
			quote = uint(sub.ModalitiesQuote.ChattingQuote.Int64)
			quoteEntity.ChattingQuote = &quote
		}
	}

	if sub.ModalitiesQuote != nil && sub.ModalitiesQuote.TextContentQuote.Valid {
		var quote uint
		if sub.ModalitiesQuote.TextContentQuote.Int64 < 0 {
			quote = 0
			quoteEntity.TextContentQuote = &quote
		} else {
			quote = uint(sub.ModalitiesQuote.TextContentQuote.Int64)
			quoteEntity.TextContentQuote = &quote
		}
	}

	if sub.ModalitiesQuote != nil && sub.ModalitiesQuote.MediaQuote.Valid {
		var quote uint
		if sub.ModalitiesQuote.MediaQuote.Int64 < 0 {
			quote = 0
			quoteEntity.MediaQuote = &quote
		} else {
			quote = uint(sub.ModalitiesQuote.MediaQuote.Int64)
			quoteEntity.MediaQuote = &quote
		}
	}

	var price string
	if sub.Price.Valid {
		price = sub.Price.String
	}

	subEntity := &entities.Subscription{
		ID:              sub.ID,
		Name:            sub.Name,
		ModalitiesQuote: quoteEntity,
		Period:          entities.Period(sub.Period),
		Price:           price,
		Level:           uint(sub.Level),
		CreatedAt:       sub.CreatedAt,
		UpdatedAt:       sub.UpdatedAt,
	}

	return subEntity, nil
}

func (d *Database) createDevBaseSubscription(ctx context.Context) error {
	const op = "infrastructure.sql.postgres.createDevBaseSubscription"

	q := `insert into subscriptions (id, name, modalities_quote, period, price, level) VALUES ($1, $2, $3, $4, $5, $6)`

	_, err := d.pool.Exec(ctx, q, uuid.New(), "base", nil, "day", "0.00", 0)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)

	}

	return nil
}
