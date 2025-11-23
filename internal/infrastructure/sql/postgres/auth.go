package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/domain"
	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/domain/aggregates"
	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/domain/entities"
	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/infrastructure/sql/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (d *Database) GetUserByID(ctx context.Context, id uuid.UUID) (*aggregates.User, error) {
	const op = "infrastructure.sql.postgres.GetUserByID"

	q := `select 
u.id as user_id, 
u.email as user_email, 
u.avatar_url as user_avatar_url, 
u.name as user_name, 
u.last_name as user_last_name, 
u.theme as user_theme, 
u.google_access_token as user_google_access_token, 
u.google_refresh_token as user_google_refresh_token, 
u.created_at as user_created_at, 
u.updated_at as user_updated_at,
s.id as session_id,
s.token as session_token,
s.expires_at as session_expires_at,
s.last_login as session_last_login,
s.user_id as session_user_id,
s.device as session_device,
s.unlogged_user_id as session_unlogged_user_id,
s.created_at as session_created_at,
s.updated_at as session_updated_at,
sub.name as plan_name,
sub.level as plan_level,
p.id as plan_id,
p.modalities_quote as plan_modalities_quote,
p.user_id as plan_user_id,
p.subscription_id as plan_subscription_id,
p.created_at as plan_created_at,
p.updated_at as plan_updated_at,
m.id as model_id,
m.name as model_name,
m.modalities as model_modalities,
m.min_level as model_min_level,
m.created_at as model_created_at,
m.updated_at as model_updated_at
from users as u left join sessions as s on u.id = s.user_id left join plans as p on u.id = p.user_id 
left join subscriptions as sub on p.subscription_id = sub.id left join models as m on sub.level >= m.min_level 
where u.id = $1`

	rows, err := d.pool.Query(ctx, q, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.NewNotFound(err, "user not found", "id", id.String())
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	type userDTO struct {
		models.User
		models.Session
		models.Plan
		models.Model
	}

	dtos, err := pgx.CollectRows(rows, pgx.RowToStructByName[userDTO])
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if len(dtos) == 0 {
		return nil, domain.NewNotFound(nil, "user not found", "id", id.String())
	}

	mp := make(map[uuid.UUID]aggregates.User)
	modelEntity := make(map[uuid.UUID]*entities.Model)
	planEntity := make(map[uuid.UUID]*entities.Plan)
	sessionsEntity := make(map[uuid.UUID]*entities.Session)
	var userID uuid.UUID
	for _, dto := range dtos {
		// Because the email address in the user table is unique,
		// you can be sure that the userID will be the only one.
		userID = dto.User.ID

		us, ok := mp[dto.User.ID]
		if !ok {
			u := &entities.User{
				ID:                dto.User.ID,
				Email:             dto.User.Email,
				AvatarURL:         dto.User.AvatarURL,
				Name:              dto.User.Name,
				Theme:             entities.Theme(dto.User.Theme),
				GoogleAccessToken: dto.User.GoogleAccessToken,
				CreatedAt:         dto.User.CreatedAt,
				UpdatedAt:         dto.User.UpdatedAt,
			}

			if dto.User.LastName.Valid {
				u.LastName = dto.User.LastName.String
			}

			if dto.User.GoogleRefreshToken.Valid {
				u.GoogleRefreshToken = dto.User.GoogleRefreshToken.String
			}
			us.User = u
			mp[dto.User.ID] = us
		}

		pl, ok := planEntity[dto.Plan.ID]
		if !ok {
			pl = &entities.Plan{
				ID:             dto.Plan.ID,
				SubscriptionID: dto.Plan.SubscriptionID,
				Level:          dto.Plan.Level,
				CreatedAt:      dto.Plan.CreatedAt,
				UpdatedAt:      dto.Plan.UpdatedAt,
			}

			if dto.Plan.UserID != nil {
				pl.UserID = *dto.Plan.UserID
			}

			qoute := entities.ModalitiesQuote{}
			if dto.Plan.ModalitiesQuote.MediaQuote.Valid {
				var mediaQuote uint = 0
				if dto.Plan.ModalitiesQuote.MediaQuote.Int64 < 0 {
					qoute.MediaQuote = &mediaQuote
				} else {
					mediaQuote = uint(dto.Plan.ModalitiesQuote.MediaQuote.Int64)
				}
				qoute.MediaQuote = &mediaQuote
			}

			if dto.Plan.ModalitiesQuote.ChattingQuote.Valid {
				var chattingQuote uint = 0
				if dto.Plan.ModalitiesQuote.ChattingQuote.Int64 < 0 {
					qoute.ChattingQuote = &chattingQuote
				} else {
					chattingQuote = uint(dto.Plan.ModalitiesQuote.ChattingQuote.Int64)
				}
				qoute.ChattingQuote = &chattingQuote
			}

			if dto.Plan.ModalitiesQuote.TextContentQuote.Valid {
				var textContentQuote uint = 0
				if dto.Plan.ModalitiesQuote.TextContentQuote.Int64 < 0 {
					qoute.TextContentQuote = &textContentQuote
				} else {
					textContentQuote = uint(dto.Plan.ModalitiesQuote.TextContentQuote.Int64)
				}
				qoute.TextContentQuote = &textContentQuote
			}

			pl.ModalitiesQuote = qoute

			planEntity[dto.Plan.ID] = pl

			us.Plan = pl
		}

		session, ok := sessionsEntity[dto.Session.ID]
		if !ok {
			session = &entities.Session{
				ID:        dto.Session.ID,
				Token:     dto.Session.Token,
				ExpiresAt: dto.Session.ExpiresAt,
				UserID:    dto.Session.UserID,
				Device:    dto.Session.Device,
				CreatedAt: dto.Session.CreatedAt,
				UpdatedAt: dto.Session.UpdatedAt,
			}

			if dto.Session.UserID != nil {
				session.UserID = dto.Session.UserID
			}

			if dto.Session.UnloggedUserID != nil {
				session.UserID = dto.Session.UnloggedUserID
			}

			sessionsEntity[dto.Session.ID] = session

			us.Sessions = append(us.Sessions, session)
		}

		model, ok := modelEntity[dto.Model.ID]
		if !ok {
			mod := make([]entities.Modality, 0)
			for _, v := range dto.Model.Modalities {
				mod = append(mod, entities.Modality(v))
			}
			model = &entities.Model{
				ID:         dto.Model.ID,
				Name:       dto.Model.Name,
				Modalities: mod,
				MinLevel:   dto.Model.MinLevel,
				CreatedAt:  dto.Model.CreatedAt,
				UpdatedAt:  dto.Model.UpdatedAt,
			}

			modelEntity[dto.Model.ID] = model

			us.Models = append(us.Models, model)
		}

		mp[dto.User.ID] = us
	}

	us := mp[userID]

	return &us, nil
}

func (d *Database) updatePlan(
	ctx context.Context,
	tx pgx.Tx,
	planInDb *entities.Plan,
	plan *entities.Plan) error {
	const op = "infrastructure.sql.postgres.saveOrUpdatePlan"

	changes := make([]string, 0)
	args := make([]interface{}, 0)
	idx := 1

	if plan.ModalitiesQuote != planInDb.ModalitiesQuote {
		changes = append(changes, fmt.Sprintf("modalities_quote = $%d", idx))
		args = append(args, plan.ModalitiesQuote)
		idx++
	}

	if plan.UpdatedAt != planInDb.UpdatedAt {
		changes = append(changes, fmt.Sprintf("updated_at = $%d", idx))
		args = append(args, plan.UpdatedAt)
		idx++
	}

	idx++

	args = append(args, plan.ID)

	if len(changes) == 0 {
		return nil
	}

	q := fmt.Sprintf("UPDATE plans SET %s WHERE id = $%d", strings.Join(changes, ", "), idx)

	_, err := tx.Exec(ctx, q, args...)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (d *Database) saveOrUpdateSession(
	ctx context.Context,
	tx pgx.Tx,
	sessionsInDb []*entities.Session,
	sessions []*entities.Session) error {
	const op = "infrastructure.sql.postgres.saveOrUpdateSession"

	if len(sessionsInDb) != len(sessions) {
		mpSessions := make(map[uuid.UUID]*entities.Session)
		for _, session := range sessionsInDb {
			mpSessions[session.ID] = session
		}

		sessionsToSave := make([]*entities.Session, 0)

		for _, session := range sessions {
			_, ok := mpSessions[session.ID]
			if !ok {
				sessionsToSave = append(sessionsToSave, session)
			}
		}

		q := `insert into sessions (id, token, expires_at, last_login, device, user_id, created_at, updated_at) 
		values ($1, $2, $3, $4, $5, $6, $7, $8)`

		var bh pgx.Batch
		for _, v := range sessionsToSave {
			bh.Queue(q, v.ID, v.Token, v.ExpiresAt, v.LastLogin, v.Device, v.UserID, v.CreatedAt, v.UpdatedAt)
		}

		br := tx.SendBatch(ctx, &bh)

		for range sessionsToSave {
			_, err := br.Exec()
			if err != nil {
				br.Close()

				return fmt.Errorf("%s: %w", op, err)
			}
		}
		br.Close()
	}

	var bh pgx.Batch
	mpSessions := make(map[uuid.UUID]*entities.Session)
	mpDbSessions := make(map[uuid.UUID]*entities.Session)
	countBath := 0

	for _, session := range sessionsInDb {
		mpDbSessions[session.ID] = session
	}

	for _, session := range sessions {
		mpSessions[session.ID] = session
	}

	for id, session := range mpSessions {
		changes := make([]string, 0)
		args := make([]interface{}, 0)
		idx := 1

		sessionInDb, ok := mpDbSessions[id]
		if !ok {
			continue
		}

		if sessionInDb == nil {
			continue
		}

		if sessionInDb.Token != session.Token {
			changes = append(changes, fmt.Sprintf("token = $%d", idx))
			args = append(args, session.Token)
			idx++
		}

		if sessionInDb.ExpiresAt != session.ExpiresAt {
			changes = append(changes, fmt.Sprintf("expires_at = $%d", idx))
			args = append(args, session.ExpiresAt)
			idx++
		}

		if sessionInDb.LastLogin != session.LastLogin {
			changes = append(changes, fmt.Sprintf("last_login = $%d", idx))
			args = append(args, session.LastLogin)
			idx++
		}

		if sessionInDb.Device != session.Device {
			changes = append(changes, fmt.Sprintf("device = $%d", idx))
			args = append(args, session.Device)
			idx++
		}

		if sessionInDb.UpdatedAt != session.UpdatedAt {
			changes = append(changes, fmt.Sprintf("updated_at = $%d", idx))
			args = append(args, session.UpdatedAt)
			idx++
		}

		args = append(args, session.ID)

		if len(changes) == 0 {
			continue
		}

		q := fmt.Sprintf("UPDATE sessions SET %s WHERE id = $%d", strings.Join(changes, ", "), idx)

		bh.Queue(q, args...)

		countBath++
	}

	br := tx.SendBatch(ctx, &bh)
	defer br.Close()

	for range countBath {
		_, err := br.Exec()
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
	}

	return nil
}

func (d *Database) GetUserByEmail(ctx context.Context, email string) (*aggregates.User, error) {
	const op = "infrastructure.sql.postgres.GetUserByEmail"

	q := `select 
u.id as user_id, 
u.email as user_email, 
u.avatar_url as user_avatar_url, 
u.name as user_name, 
u.last_name as user_last_name, 
u.theme as user_theme, 
u.google_access_token as user_google_access_token, 
u.google_refresh_token as user_google_refresh_token, 
u.created_at as user_created_at, 
u.updated_at as user_updated_at,
s.id as session_id,
s.token as session_token,
s.expires_at as session_expires_at,
s.last_login as session_last_login,
s.user_id as session_user_id,
s.device as session_device,
s.unlogged_user_id as session_unlogged_user_id,
s.created_at as session_created_at,
s.updated_at as session_updated_at,
sub.name as plan_name,
sub.level as plan_level,
p.id as plan_id,
p.modalities_quote as plan_modalities_quote,
p.user_id as plan_user_id,
p.subscription_id as plan_subscription_id,
p.created_at as plan_created_at,
p.updated_at as plan_updated_at,
m.id as model_id,
m.name as model_name,
m.modalities as model_modalities,
m.min_level as model_min_level,
m.created_at as model_created_at,
m.updated_at as model_updated_at
from users as u left join sessions as s on u.id = s.user_id left join plans as p on u.id = p.user_id 
left join subscriptions as sub on p.subscription_id = sub.id left join models as m on sub.level >= m.min_level 
where email = $1`

	rows, err := d.pool.Query(ctx, q, email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("%s: %w", op, domain.NewNotFound(
				err,
				"user not found",
				"email",
				email,
			))
		}

		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	dtos, err := pgx.CollectRows(rows, pgx.RowToStructByNameLax[userDTO])
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if len(dtos) == 0 {
		return nil, domain.NewNotFound(nil, "user not found", "email", email)
	}

	mp := make(map[uuid.UUID]aggregates.User)
	planEntity := make(map[uuid.UUID]*entities.Plan)
	sessionsEntity := make(map[uuid.UUID]*entities.Session)
	modelEntity := make(map[uuid.UUID]*entities.Model)

	var userID uuid.UUID

	for _, dto := range dtos {
		// Because the email address in the user table is unique,
		// you can be sure that the userID will be the only one.
		userID = dto.User.ID

		us, ok := mp[dto.User.ID]
		if !ok {
			u := &entities.User{
				ID:                dto.User.ID,
				Email:             dto.User.Email,
				AvatarURL:         dto.User.AvatarURL,
				Name:              dto.User.Name,
				Theme:             entities.Theme(dto.User.Theme),
				GoogleAccessToken: dto.User.GoogleAccessToken,
				CreatedAt:         dto.User.CreatedAt,
				UpdatedAt:         dto.User.UpdatedAt,
			}

			if dto.User.LastName.Valid {
				u.LastName = dto.User.LastName.String
			}

			if dto.User.GoogleRefreshToken.Valid {
				u.GoogleRefreshToken = dto.User.GoogleRefreshToken.String
			}

			us.User = u

			mp[dto.User.ID] = us
		}

		pl, ok := planEntity[dto.Plan.ID]
		if !ok {
			pl = &entities.Plan{
				ID:             dto.Plan.ID,
				SubscriptionID: dto.Plan.SubscriptionID,
				Level:          dto.Plan.Level,
				CreatedAt:      dto.Plan.CreatedAt,
				UpdatedAt:      dto.Plan.UpdatedAt,
			}

			if dto.Plan.UserID != nil {
				pl.UserID = *dto.Plan.UserID
			}

			qoute := entities.ModalitiesQuote{}
			if dto.Plan.ModalitiesQuote.MediaQuote.Valid {
				var mediaQuote uint = 0
				if dto.Plan.ModalitiesQuote.MediaQuote.Int64 < 0 {
					qoute.MediaQuote = &mediaQuote
				} else {
					mediaQuote = uint(dto.Plan.ModalitiesQuote.MediaQuote.Int64)
				}
				qoute.MediaQuote = &mediaQuote
			}

			if dto.Plan.ModalitiesQuote.ChattingQuote.Valid {
				var chattingQuote uint = 0
				if dto.Plan.ModalitiesQuote.ChattingQuote.Int64 < 0 {
					qoute.ChattingQuote = &chattingQuote
				} else {
					chattingQuote = uint(dto.Plan.ModalitiesQuote.ChattingQuote.Int64)
				}
				qoute.ChattingQuote = &chattingQuote
			}

			if dto.Plan.ModalitiesQuote.TextContentQuote.Valid {
				var textContentQuote uint = 0
				if dto.Plan.ModalitiesQuote.TextContentQuote.Int64 < 0 {
					qoute.TextContentQuote = &textContentQuote
				} else {
					textContentQuote = uint(dto.Plan.ModalitiesQuote.TextContentQuote.Int64)
				}
				qoute.TextContentQuote = &textContentQuote
			}

			pl.ModalitiesQuote = qoute

			planEntity[dto.Plan.ID] = pl

			us.Plan = pl
		}

		session, ok := sessionsEntity[dto.Session.ID]
		if !ok {
			session = &entities.Session{
				ID:        dto.Session.ID,
				Token:     dto.Session.Token,
				ExpiresAt: dto.Session.ExpiresAt,
				UserID:    dto.Session.UserID,
				Device:    dto.Session.Device,
				CreatedAt: dto.Session.CreatedAt,
				UpdatedAt: dto.Session.UpdatedAt,
			}

			if dto.Session.UserID != nil {
				session.UserID = dto.Session.UserID
			}

			if dto.Session.UnloggedUserID != nil {
				session.UserID = dto.Session.UnloggedUserID
			}

			sessionsEntity[dto.Session.ID] = session

			us.Sessions = append(us.Sessions, session)
		}

		model, ok := modelEntity[dto.Model.ID]
		if !ok {
			mod := make([]entities.Modality, 0)
			for _, v := range dto.Model.Modalities {
				mod = append(mod, entities.Modality(v))
			}

			model = &entities.Model{
				ID:         dto.Model.ID,
				Name:       dto.Model.Name,
				Modalities: mod,
				MinLevel:   dto.Model.MinLevel,
				CreatedAt:  dto.Model.CreatedAt,
				UpdatedAt:  dto.Model.UpdatedAt,
			}

			modelEntity[dto.Model.ID] = model

			us.Models = append(us.Models, model)
		}

		mp[dto.User.ID] = us
	}

	us := mp[userID]

	return &us, nil
}
