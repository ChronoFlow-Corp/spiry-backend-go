package auth

import (
	"context"
	"errors"
	"fmt"

	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/domain"
	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/domain/aggregates"
	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/domain/entities"
	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/infrastructure/sql/storages/models"
	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/infrastructure/sql/storages/plans"
	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/infrastructure/sql/storages/sessions"
	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/infrastructure/sql/storages/subscriptions"
	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/infrastructure/sql/storages/users"
	"github.com/avito-tech/go-transaction-manager/trm/v2/manager"
	"github.com/google/uuid"
)

type Repository struct {
	plan         plans.PlanStorage
	session      sessions.SessionStorage
	user         users.UserStorage
	model        models.ModelStorage
	subscription subscriptions.SubscriptionStorage
	manager      *manager.Manager
}

func NewAuthRepository(
	plan plans.PlanStorage,
	session sessions.SessionStorage,
	user users.UserStorage,
	model models.ModelStorage,
	manager *manager.Manager,
	subscription subscriptions.SubscriptionStorage,
) *Repository {
	return &Repository{
		plan:         plan,
		session:      session,
		user:         user,
		model:        model,
		subscription: subscription,
		manager:      manager,
	}
}

func (r *Repository) SaveUser(ctx context.Context, userAggregate *aggregates.User) error {
	const op = "sql.repository.auth.SaveUser"
	err := r.manager.Do(ctx, func(ctx context.Context) error {
		_, err := r.user.GetByEmail(ctx, userAggregate.Email)
		if err != nil {
			if errors.Is(err, users.ErrNotFound) {
				err = r.user.Create(ctx, *userAggregate.User)
				if err != nil {
					return err
				}
			} else {
				return err
			}
		}

		err = r.user.Update(ctx, *userAggregate.User)
		if err != nil {
			return err
		}

		_, err = r.plan.GetByID(ctx, userAggregate.Plan.ID)
		if err != nil {
			if errors.Is(err, plans.ErrNotFound) {
				err = r.plan.Create(ctx, *userAggregate.Plan)
				if err != nil {
					return err
				}
			} else {
				return err
			}
		}

		err = r.plan.Update(ctx, *userAggregate.Plan)
		if err != nil {
			return err
		}

		s, err := r.session.GetByUserID(ctx, userAggregate.ID)
		if err != nil {
			return err
		}

		mpSessions := make(map[uuid.UUID]*entities.Session)
		for _, ss := range s {
			mpSessions[ss.ID] = ss
		}

		for _, ss := range userAggregate.Sessions {
			if _, ok := mpSessions[ss.ID]; !ok {
				err = r.session.Create(ctx, *ss)
				if err != nil {
					return err
				}
			} else {
				err = r.session.Update(ctx, *ss)
				if err != nil {
					return err
				}
			}
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (r *Repository) GetUserByEmail(ctx context.Context, email string) (*aggregates.User, error) {
	const op = "sql.repository.auth.GetUserByEmail"

	us, err := r.user.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, users.ErrNotFound) {
			return nil, domain.NewNotFound(err, "user not found", "email", email)
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	userAggregate, err := r.aggregateFromUser(ctx, us)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return userAggregate, nil
}

func (r *Repository) GetUserByID(ctx context.Context, id uuid.UUID) (*aggregates.User, error) {
	const op = "sql.repository.auth.GetUserByID"

	us, err := r.user.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	userAggregate, err := r.aggregateFromUser(ctx, us)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return userAggregate, nil
}

func (r *Repository) GetBaseSubscription(ctx context.Context) (*entities.Subscription, error) {
	const op = "sql.repository.auth.GetBaseSubscription"

	subs, err := r.subscription.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	//TODO: change to proper base subscription selection
	return &subs[0], nil
}

func (r *Repository) GetModels(ctx context.Context) ([]*entities.Model, error) {
	const op = "sql.repository.auth.GetModels"

	modelsList, err := r.model.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	m := make([]*entities.Model, len(modelsList))
	for _, model := range modelsList {
		m = append(m, &model)
	}

	return m, nil
}

func (r *Repository) GetModelByName(ctx context.Context, name string) (*entities.Model, error) {
	const op = "sql.repository.auth.GetModelByName"

	model, err := r.model.GetByName(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return model, nil
}

func (r *Repository) GetSubscriptionByID(
	ctx context.Context,
	id uuid.UUID,
) (*entities.Subscription, error) {
	const op = "sql.repository.auth.GetSubscriptionByID"

	sub, err := r.subscription.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return sub, nil
}

func (r *Repository) aggregateFromUser(
	ctx context.Context,
	us *entities.User,
) (*aggregates.User, error) {
	const op = "sql.repository.auth.aggregateFromUser"

	plan, err := r.plan.GetByUserID(ctx, us.ID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	sub, err := r.subscription.GetByID(ctx, plan.SubscriptionID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	plan.Level = sub.Level

	sessionsList, err := r.session.GetByUserID(ctx, us.ID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	modelsList, err := r.model.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	allowedModels := make([]*entities.Model, 0)

	for _, m := range modelsList {
		if plan.Level >= m.MinLevel {
			allowedModels = append(allowedModels, &m)
		}
	}

	userAggregate, err := aggregates.NewUser(us, plan, allowedModels, sessionsList)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return userAggregate, nil
}
