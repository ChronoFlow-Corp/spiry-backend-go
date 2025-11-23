package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/domain"
	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/domain/aggregates"
	"github.com/jackc/pgx/v5"
)

func (d *Database) SaveUser(ctx context.Context, user *aggregates.User) error {
	const op = "infrastructure.sql.postgres.SaveUser"

	tx, err := d.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	defer func() {
		if err != nil {
			txErr := tx.Rollback(ctx)
			if txErr != nil {
				err = fmt.Errorf("tx err: %w, rollback err: %w", err, txErr)

				return
			}
		}

		txErr := tx.Commit(ctx)
		if txErr != nil {
			err = fmt.Errorf("commit err: %w", txErr)
		}
	}()

	userInDatabase, err := d.GetUserByEmail(ctx, user.Email)
	if err != nil {
		var notFound *domain.ErrorNotFound
		if errors.As(err, &notFound) {
			err = d.saveNewUser(ctx, user)
			if err != nil {
				return fmt.Errorf("%s: %w", op, err)
			}

			return nil
		}

		return fmt.Errorf("%s: %w", op, err)
	}

	changes := make([]string, 0)
	args := make([]interface{}, 0)
	idx := 1

	if user.AvatarURL != userInDatabase.AvatarURL {
		changes = append(changes, fmt.Sprintf("avatar_url = $%d", idx))
		args = append(args, user.AvatarURL)
		idx++
	}

	if user.Name != userInDatabase.Name {
		changes = append(changes, fmt.Sprintf("name = $%d", idx))
		args = append(args, user.Name)
		idx++
	}

	if user.LastName != userInDatabase.LastName {
		changes = append(changes, fmt.Sprintf("last_name = $%d", idx))
		args = append(args, user.LastName)
		idx++
	}

	if user.Theme != userInDatabase.Theme {
		changes = append(changes, fmt.Sprintf("theme = $%d", idx))
		args = append(args, user.Theme)
		idx++
	}

	if user.GoogleAccessToken != userInDatabase.GoogleAccessToken {
		changes = append(changes, fmt.Sprintf("google_access_token = $%d", idx))
		args = append(args, user.GoogleAccessToken)
		idx++
	}

	if user.GoogleRefreshToken != userInDatabase.GoogleRefreshToken {
		changes = append(changes, fmt.Sprintf("google_refresh_token = $%d", idx))
		args = append(args, user.GoogleRefreshToken)
		idx++
	}

	if user.UpdatedAt != userInDatabase.UpdatedAt {
		changes = append(changes, fmt.Sprintf("updated_at = $%d", idx))
		args = append(args, user.UpdatedAt)
		idx++
	}

	if len(changes) != 0 {
		// Email
		idx++
		args = append(args, user.User.Email)

		q := fmt.Sprintf("UPDATE users SET %s WHERE email = $%d", strings.Join(changes, ", "), idx)

		_, err = d.pool.Exec(ctx, q, args...)
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
	}

	err = d.saveOrUpdateSession(ctx, tx, userInDatabase.Sessions, user.Sessions)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	err = d.updatePlan(ctx, tx, userInDatabase.Plan, user.Plan)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (d *Database) saveNewUser(ctx context.Context, user *aggregates.User) error {
	const op = "infrastructure.sql.postgres.saveNewUser"

	tx, err := d.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	defer func() {
		if err != nil {
			txErr := tx.Rollback(ctx)
			if txErr != nil {
				err = fmt.Errorf("tx err: %w, rollback err: %w", err, txErr)

				return
			}
		}

		txErr := tx.Commit(ctx)
		if txErr != nil {
			err = fmt.Errorf("commit err: %w", txErr)
		}
	}()

	q := `insert into users (id, 
	email, 
	avatar_url,
	name,
	last_name, 
	google_access_token,
	google_refresh_token,
	created_at,
	updated_at) 
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`

	_, err = tx.Exec(
		ctx,
		q,
		user.ID,
		user.Email,
		user.AvatarURL,
		user.Name,
		user.LastName,
		user.GoogleAccessToken,
		user.GoogleRefreshToken,
		user.CreatedAt,
		user.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	q = `insert into plans(id, modalities_quote, user_id, subscription_id, created_at, updated_at) 
	VALUES ($1, $2, $3, $4, $5, $6)`

	_, err = tx.Exec(
		ctx,
		q,
		user.Plan.ID,
		user.Plan.ModalitiesQuote,
		user.Plan.UserID,
		user.Plan.SubscriptionID,
		user.Plan.CreatedAt,
		user.Plan.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	q = `insert into 
sessions(id, token, expires_at, device, last_login, user_id, unlogged_user_id, created_at, updated_at) 
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`

	var batch pgx.Batch
	for _, v := range user.Sessions {
		batch.Queue(q, v.ID, v.Token, v.ExpiresAt, v.Device, v.LastLogin, v.UserID, nil, v.CreatedAt, v.UpdatedAt)
	}

	br := tx.SendBatch(ctx, &batch)
	defer br.Close()

	for range user.Sessions {
		_, err = br.Exec()
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
	}

	return nil
}
