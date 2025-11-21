package pgx

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/domain/entities"
	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/infrastructure/sql/storages/plans"
	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/infrastructure/sql/storages/plans/models"
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
var _ plans.PlanStorage = (*Pgx)(nil)

func NewPgx(pool *pgxpool.Pool) *Pgx {
	return &Pgx{
		pool:   pool,
		getter: trmgr.DefaultCtxGetter,
	}
}

func (p *Pgx) Create(ctx context.Context, plan entities.Plan) error {
	const op = "storages.plans.pgx.Create"

	conn := p.getter.DefaultTrOrDB(ctx, p.pool)

	query, values, err := sq.
		Insert(table).
		Columns(columns...).
		Values(
			plan.ID,
			plan.ModalitiesQuote,
			plan.UserID,
			plan.SubscriptionID,
			plan.CreatedAt,
			plan.UpdatedAt,
		).ToSql()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	_, err = conn.Exec(ctx, query, values...)
	if err != nil {
		return handlePgxError(op, err)
	}

	return nil
}

func (p *Pgx) GetByUserID(ctx context.Context, userID uuid.UUID) (*entities.Plan, error) {
	const op = "storages.plans.pgx.GetByUserID"

	conn := p.getter.DefaultTrOrDB(ctx, p.pool)

	query, values, err := sq.
		Select(columns...).
		From(table).
		Where(squirrel.Eq{"user_id": userID}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	row := conn.QueryRow(ctx, query, values...)

	plan, err := scanToEntity(row)
	if err != nil {

		return nil, handlePgxError(op, err)
	}

	return &plan, nil
}

func (p *Pgx) GetByID(ctx context.Context, planID uuid.UUID) (*entities.Plan, error) {
	const op = "storages.plans.pgx.GetByID"

	conn := p.getter.DefaultTrOrDB(ctx, p.pool)

	query, values, err := sq.
		Select(columns...).
		From(table).
		Where(squirrel.Eq{columns[id]: planID}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	row := conn.QueryRow(ctx, query, values...)

	plan, err := scanToEntity(row)
	if err != nil {
		return nil, handlePgxError(op, err)
	}

	return &plan, nil
}

func (p *Pgx) Update(ctx context.Context, plan entities.Plan) error {
	const op = "storages.plans.pgx.Update"

	conn := p.getter.DefaultTrOrDB(ctx, p.pool)

	planDB, err := p.GetByID(ctx, plan.ID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	update := sq.Update(table).Where(squirrel.Eq{columns[id]: plan.ID})

	if plan.ModalitiesQuote != planDB.ModalitiesQuote {
		update = update.Set(columns[modalitiesQuote], plan.ModalitiesQuote)
	}

	if plan.SubscriptionID != planDB.SubscriptionID {
		update = update.Set(columns[subscriptionID], plan.SubscriptionID)
	}

	update = update.Set(columns[updatedAt], plan.UpdatedAt)

	query, values, err := update.ToSql()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	_, err = conn.Exec(ctx, query, values...)
	if err != nil {
		return handlePgxError(op, err)
	}

	return nil
}

func scanToEntity(row pgx.Row) (entities.Plan, error) {
	var p models.Plan
	err := row.Scan(
		&p.ID,
		&p.ModalitiesQuote,
		&p.UserID,
		&p.SubscriptionID,
		&p.CreatedAt,
		&p.UpdatedAt,
	)
	if err != nil {
		return entities.Plan{}, err
	}

	plan := entities.Plan{
		ID:             p.ID,
		SubscriptionID: p.SubscriptionID,
		Level:          p.Level,
		UserID:         p.UserID,
		CreatedAt:      p.CreatedAt,
		UpdatedAt:      p.UpdatedAt,
	}

	var mod models.ModalitiesQuote
	err = json.Unmarshal(p.ModalitiesQuote, &mod)
	if err != nil {
		return entities.Plan{}, err
	}

	plan.ModalitiesQuote = entities.ModalitiesQuote{
		ChattingQuote:    mod.ChattingQuote,
		MediaQuote:       mod.MediaQuote,
		TextContentQuote: mod.TextContentQuote,
	}

	return plan, nil
}
