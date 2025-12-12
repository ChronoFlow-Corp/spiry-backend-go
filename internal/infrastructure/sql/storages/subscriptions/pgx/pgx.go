package pgx

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/domain/entities"
	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/infrastructure/sql/storages/subscriptions"
	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/infrastructure/sql/storages/subscriptions/models"
	"github.com/Masterminds/squirrel"
	trmgr "github.com/avito-tech/go-transaction-manager/drivers/pgxv5/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Pgx struct {
	pool   *pgxpool.Pool
	getter *trmgr.CtxGetter
}

var sq = squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
var _ subscriptions.SubscriptionStorage = (*Pgx)(nil)

func NewPgx(pool *pgxpool.Pool) *Pgx {
	return &Pgx{
		pool:   pool,
		getter: trmgr.DefaultCtxGetter,
	}
}

func (p *Pgx) Create(ctx context.Context, subscription entities.Subscription) error {
	const op = "storages.subscriptions.pgx.Create"

	conn := p.getter.DefaultTrOrDB(ctx, p.pool)

	query, values, err := sq.
		Insert(table).
		Columns(columns...).
		Values(
			subscription.ID,
			subscription.Name,
			subscription.Quote,
			subscription.Period,
			subscription.Price,
			subscription.Level,
			subscription.CreatedAt,
			subscription.UpdatedAt,
		).ToSql()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	_, err = conn.Exec(ctx, query, values...)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (p *Pgx) GetByID(ctx context.Context, subID uuid.UUID) (*entities.Subscription, error) {
	const op = "storages.subscriptions.pgx.GetByID"

	conn := p.getter.DefaultTrOrDB(ctx, p.pool)

	query, values, err := sq.
		Select(columns...).
		From(table).
		Where(squirrel.Eq{columns[id]: subID}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	row := conn.QueryRow(ctx, query, values...)

	sub, err := scanToEntity(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("%s: %w: %w", op, subscriptions.ErrNotFound, err)
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &sub, nil
}

func (p *Pgx) GetAll(ctx context.Context) ([]entities.Subscription, error) {
	const op = "storages.subscriptions.pgx.GetAll"

	conn := p.getter.DefaultTrOrDB(ctx, p.pool)

	query, values, err := sq.
		Select(columns...).
		From(table).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	rows, err := conn.Query(ctx, query, values...)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	var subs []entities.Subscription

	for rows.Next() {
		sub, err := scanToEntity(rows)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, fmt.Errorf("%s: %w: %w", op, subscriptions.ErrNotFound, err)
			}
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		subs = append(subs, sub)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return subs, nil
}

func scanToEntity(row pgx.Row) (entities.Subscription, error) {
	var sub models.Subscription
	err := row.Scan(
		&sub.ID,
		&sub.Name,
		&sub.Quote,
		&sub.Period,
		&sub.Price,
		&sub.Level,
		&sub.CreatedAt,
		&sub.UpdatedAt,
	)
	if err != nil {
		return entities.Subscription{}, err
	}

	var mod models.Quote

	if len(sub.Quote) > 0 {
		err = json.Unmarshal(sub.Quote, &mod)
		if err != nil {
			return entities.Subscription{}, err
		}
	}

	subEntity := entities.Subscription{
		ID:        sub.ID,
		Name:      sub.Name,
		Period:    entities.Period(sub.Period),
		Level:     uint(sub.Level),
		CreatedAt: sub.CreatedAt,
		UpdatedAt: sub.UpdatedAt,
	}

	if sub.Price.Valid {
		subEntity.Price = sub.Price.String
	}

	subEntity.Quote = entities.Quote{
		ToolLimits:       make([]entities.ToolLimit, len(mod.ToolLimits)),
		MediaLimit:       make([]entities.MediaLimit, len(mod.MediaLimit)),
		FlagLimits:       make([]entities.FlagLimit, len(mod.FlagLimits)),
		ResetQuotePeriod: entities.Period(mod.ResetQuotePeriod),
	}

	for i, limit := range mod.ToolLimits {
		subEntity.Quote.ToolLimits[i] = entities.ToolLimit{
			ID:            limit.ID,
			SettingsLimit: limit.SettingsLimit,
			Usage:         limit.Usage,
		}
	}

	for i, limit := range mod.MediaLimit {
		subEntity.Quote.MediaLimit[i] = entities.MediaLimit{
			Type:     limit.Type,
			Upload:   limit.Upload,
			Generate: limit.Generate,
			Size:     uint64(limit.Size),
		}
	}

	for i, limit := range mod.FlagLimits {
		subEntity.Quote.FlagLimits[i] = entities.FlagLimit{
			Name: limit.Name,
		}
	}

	return subEntity, nil
}
