package postgres

import (
	"context"
	"fmt"
	"net/url"

	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/domain"
	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/domain/aggregates"
	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/domain/entities"
	models2 "github.com/ChronoFlow-Corp/spiry-backend-go/internal/domain/models"
	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/infrastructure/sql/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (d *Database) SaveCommand(ctx context.Context, cm *aggregates.Command) error {
	const op = "infrastructure.sql.postgres.chatting.SaveCommand"

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

	err = d.saveNewCommand(ctx, tx, cm)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	q := `insert into command_medias
	(id, name, type, url, size, command_id, user_id, created_at, updated_at) VALUES
	($1, $2, $3, $4, $5, $6, $7, $8, $9)`

	var bh pgx.Batch

	for _, m := range cm.Medias {
		bh.Queue(
			q,
			m.ID,
			m.Name,
			m.Type,
			m.URL,
			m.Size,
			m.CommandID,
			m.UserID,
			m.CreatedAt,
			m.UpdatedAt,
		)
	}

	br := tx.SendBatch(ctx, &bh)

	defer br.Close()

	for range cm.Medias {
		_, err = br.Exec()
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
	}

	return nil
}

func (d *Database) GetCommandByID(ctx context.Context, id uuid.UUID) (*aggregates.Command, error) {
	const op = "infrastructure.sql.postgres.chatting.GetCommandByID"

	q := `select
c.id as command_id,
c.prompt as command_prompt,
c.settings as command_settings,
c.flags as command_flags,
c.chat_id as command_chat_id,
c.tool_id as command_tool_id,
c.model_id as command_model_id,
c.status as command_status,
c.user_id as command_user_id,
c.unlogged_user_id as command_unlogged_user_id,
c.created_at as command_created_at,
c.updated_at as command_updated_at,
m.id as model_id,
m.name as model_name,
m.modalities as model_modalities,
m.min_level as model_min_level,
m.created_at as model_created_at,
m.updated_at as model_updated_at,
t.id as tool_id,
t.name as tool_name,
t.modalities as tool_modalities,
t.settings as tool_settings,
t.prompt as tool_prompt,
t.min_level as tool_min_level,
t.created_at as tool_created_at,
t.updated_at as tool_updated_at,
cm.id as command_media_id,
cm.name as command_media_name,
cm.type as command_media_type,
cm.url as command_media_url,
cm.size as command_media_size,
cm.command_id as command_media_command_id,
cm.user_id as command_media_user_id,
cm.unlogged_user_id as command_media_unlogged_user_id,
cm.created_at as command_media_created_at,
cm.updated_at as command_media_updated_at
from commands as c 
left join models as m on c.model_id = m.id 
left join tools as t on c.tool_id = t.id
left join command_medias as cm on cm.command_id = c.id
where c.id = $1`

	rows, err := d.pool.Query(ctx, q, id)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	defer rows.Close()

	type dbCommand struct {
		models.Command
		models.Model
		models.Tool
		models.CommandMedia
	}

	dtos, err := pgx.CollectRows(rows, pgx.RowToStructByName[dbCommand])
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if len(dtos) == 0 {
		return nil, domain.NewNotFound(nil, "command not found", "id", id.String())
	}

	mpCommand := make(map[uuid.UUID]*entities.Command)
	mpTool := make(map[uuid.UUID]*entities.Tool)
	mpModel := make(map[uuid.UUID]*entities.Model)
	mpCommandMedia := make(map[uuid.UUID]*entities.CommandMedia)

	commandAggregate := &aggregates.Command{
		Medias: make([]*entities.CommandMedia, 0),
	}

	for _, dto := range dtos {
		com, ok := mpCommand[dto.Command.ID]
		if !ok {
			var userID uuid.UUID
			if dto.Command.UnloggedUserID != uuid.Nil {
				userID = dto.Command.UnloggedUserID
			}
			if dto.Command.UserID != uuid.Nil {
				userID = dto.Command.UserID
			}

			com = &entities.Command{
				ID:        dto.Command.ID,
				Prompt:    dto.Command.Prompt,
				Settings:  dto.Command.Settings,
				Flags:     dto.Command.Flags,
				ChatID:    dto.Command.ChatID,
				ToolID:    &dto.Command.ToolID,
				ModelID:   dto.Command.ModelID,
				Status:    dto.Command.Status,
				UserID:    userID,
				CreatedAt: dto.Command.CreatedAt,
				UpdatedAt: dto.Command.UpdatedAt,
			}

			mpCommand[dto.Command.ID] = com

			commandAggregate.Command = com
		}

		tool, ok := mpTool[dto.Tool.ID]
		if !ok {
			mod := make([]entities.Modality, 0, len(dto.Tool.Modalities))
			for _, m := range dto.Tool.Modalities {
				mod = append(mod, entities.Modality(m))
			}
			tool = &entities.Tool{
				ID:         dto.Tool.ID,
				Name:       dto.Tool.Name,
				Modalities: mod,
				Settings:   dto.Tool.Settings,
				Prompt:     dto.Tool.Prompt,
				MinLevel:   uint(dto.Tool.MinLevel),
				CreatedAt:  dto.Tool.CreatedAt,
				UpdatedAt:  dto.Tool.UpdatedAt,
			}

			mpTool[dto.Tool.ID] = tool
			commandAggregate.Tool = tool
		}

		model, ok := mpModel[dto.Model.ID]
		if !ok {
			mod := make([]entities.Modality, 0, len(dto.Model.Modalities))
			for _, m := range dto.Model.Modalities {
				mod = append(mod, entities.Modality(m))
			}

			model = &entities.Model{
				ID:         dto.Model.ID,
				Name:       dto.Model.Name,
				Modalities: mod,
				MinLevel:   dto.Model.MinLevel,
				CreatedAt:  dto.Model.CreatedAt,
				UpdatedAt:  dto.Model.UpdatedAt,
			}

			mpModel[dto.Model.ID] = model

			commandAggregate.Model = model
		}

		media, ok := mpCommandMedia[dto.CommandMedia.ID]
		if !ok {
			u, err := url.Parse(dto.CommandMedia.URL)
			if err != nil {
				return nil, fmt.Errorf("%s: %w", op, err)
			}

			var userID uuid.UUID
			if dto.CommandMedia.UserID != uuid.Nil {
				userID = dto.CommandMedia.UserID
			}

			if dto.CommandMedia.UnloggedUserID != uuid.Nil {
				userID = dto.CommandMedia.UnloggedUserID
			}

			media = &entities.CommandMedia{
				ID:        dto.CommandMedia.ID,
				Name:      dto.CommandMedia.Name,
				Type:      dto.CommandMedia.Type,
				URL:       *u,
				Size:      uint64(dto.CommandMedia.Size),
				CommandID: dto.CommandMedia.CommandID,
				UserID:    userID,
				CreatedAt: dto.CommandMedia.CreatedAt,
				UpdatedAt: dto.CommandMedia.UpdatedAt,
			}

			mpCommandMedia[dto.CommandMedia.ID] = media

			commandAggregate.Medias = append(commandAggregate.Medias, media)
		}
	}

	return commandAggregate, nil
}

func (d *Database) GetChatByID(ctx context.Context, chatID uuid.UUID) (*aggregates.Chat, error) {
	const op = "infrastructure.sql.postgres.chatting.GetChatByID"

	userID, ok := ctx.Value(models2.UserIDCtxKey{}).(uuid.UUID)
	if !ok {
		return nil, fmt.Errorf("%s: %w", op, domain.NewNotFound(
			nil,
			"chats not found",
			"id",
			chatID.String(),
		))
	}

	q := `with command_medias_agg as (
    select
        cmm.command_id,
        coalesce(
                        jsonb_agg(
                        jsonb_strip_nulls(
                                jsonb_build_object(
                                        'id', cmm.id,
                                        'name', cmm.name,
                                        'type', cmm.type,
                                        'url', cmm.url,
                                        'size', cmm.size,
                                        'command_id', cmm.command_id,
                                        'user_id', cmm.user_id,
                                        'unlogged_user_id', cmm.unlogged_user_id,
                                        'created_at', cmm.created_at,
                                        'updated_at', cmm.updated_at
                                )
                        )
                                 ) filter (where cmm.id is not null),
                        '[]'::jsonb
        ) as command_medias
    from command_medias cmm
    group by cmm.command_id
),
     result_medias_agg as (
         select
             rm.result_id,
             coalesce(
                             jsonb_agg(
                             jsonb_strip_nulls(
                                     jsonb_build_object(
                                             'id', rm.id,
                                             'name', rm.name,
                                             'type', rm.type,
                                             'url', rm.url,
                                             'size', rm.size,
                                             'result_id', rm.result_id,
                                             'user_id', rm.user_id,
                                             'unlogged_user_id', rm.unlogged_user_id,
                                             'created_at', rm.created_at,
                                             'updated_at', rm.updated_at
                                     )
                             )
                                      ) filter (where rm.id is not null),
                             '[]'::jsonb
             ) as result_medias
         from result_medias rm
         group by rm.result_id
     ),
     result_per_command as (
         select
             r.command_id,
             coalesce(
                             jsonb_agg(
                             jsonb_strip_nulls(
                                     jsonb_build_object(
                                             'id', r.id,
                                             'text', r.text,
                                             'command_id', r.command_id,
                                             'tool_id', r.tool_id,
                                             'chat_id', r.chat_id,
                                             'user_id', r.user_id,
                                             'unlogged_user_id', r.unlogged_user_id,
                                             'model_id', r.model_id,
                                             'created_at', r.created_at,
                                             'updated_at', r.updated_at,
                                             'medias', coalesce(rm_agg.result_medias, '[]'::jsonb)
                                     )
                             )
                                      ) filter (where r.id is not null),
                             '[]'::jsonb
             ) as results  -- массив результатов для команды
         from results r
                  left join result_medias_agg rm_agg on rm_agg.result_id = r.id
         group by r.command_id
     )
select
    c.id as chat_id,
    c.title as chat_title,
    c.created_at as chat_created_at,
    c.updated_at as chat_updated_at,
    c.user_id as chat_user_id,
    c.unlogged_user_id as chat_unlogged_user_id,
    coalesce(
                    jsonb_agg(
                    jsonb_strip_nulls(
                            jsonb_build_object(
                                    'command', jsonb_build_object(
                                    'id', cm.id,
                                    'prompt', cm.prompt,
                                    'settings', cm.settings,
                                    'flags', cm.flags,
                                    'chat_id', cm.chat_id,
                                    'tool_id', cm.tool_id,
                                    'model_id', cm.model_id,
                                    'status', cm.status,
                                    'user_id', cm.user_id,
                                    'unlogged_user_id', cm.unlogged_user_id,
                                    'created_at', cm.created_at,
                                    'updated_at', cm.created_at,
                                    'medias', coalesce(cmm_agg.command_medias, '[]'::jsonb)
                                               ),
                                    'results', coalesce(rpc.results, '[]'::jsonb)  -- массив результатов для этой команды (пара)
                            )
                    )
                             ) filter (where cm.id is not null),
                    '[]'::jsonb
    ) as command_result_pairs 
from chats c
         left join commands cm on cm.chat_id = c.id
         left join command_medias_agg cmm_agg on cmm_agg.command_id = cm.id
         left join result_per_command rpc on rpc.command_id = cm.id
where c.id = $1
  and c.user_id = $2
group by
    c.id, c.title, c.created_at, c.updated_at, c.user_id, c.unlogged_user_id`

	rows, err := d.pool.Query(ctx, q, chatID, userID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	var chat chatDto
	for rows.Next() {
		err = rows.Scan(
			&chat.ID,
			&chat.Title,
			&chat.CreatedAt,
			&chat.UpdatedAt,
			&chat.UserID,
			&chat.UnloggedUserID,
			&chat.Couples,
		)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
	}

	chatEntity := &entities.Chat{
		ID:        chat.ID,
		Title:     chat.Title,
		CreatedAt: chat.CreatedAt,
		UpdatedAt: chat.UpdatedAt,
		UserID:    chat.UserID,
	}

	couples := make([]aggregates.CommandResultCouple, 0, len(chat.Couples))
	for _, c := range chat.Couples {
		cAgg := aggregates.CommandResultCouple{
			Command: &entities.Command{
				ID:        c.Command.ID,
				Prompt:    c.Command.Prompt,
				Settings:  c.Command.Settings,
				Flags:     c.Command.Flags,
				ChatID:    chat.ID,
				ToolID:    &c.Command.ToolID,
				ModelID:   c.Command.ModelID,
				Status:    c.Command.Status,
				UserID:    c.Command.UserID,
				CreatedAt: c.Command.CreatedAt,
				UpdatedAt: c.Command.UpdatedAt,
			},
			CommandMedia: make([]*entities.CommandMedia, 0, len(c.Command.Medias)),
			Result: &entities.Result{
				ID:           c.Result.ID,
				Text:         c.Result.Text,
				OpenRouterID: c.Result.OpenRouterID,
				CommandID:    c.Result.CommandID,
				ToolID:       c.Result.ToolID,
				ChatID:       c.Result.ChatID,
				UserID:       c.Result.UserID,
				ModelID:      c.Result.ModelID,
				CreatedAt:    c.Result.CreatedAt,
				UpdatedAt:    c.Result.UpdatedAt,
			},
			ResultMedia: make([]*entities.ResultMedia, 0, len(c.Result.Medias)),
		}

		for _, m := range c.Command.Medias {
			u, err := url.Parse(m.URL)
			if err != nil {
				return nil, fmt.Errorf("%s: %w", op, err)
			}
			cAgg.CommandMedia = append(cAgg.CommandMedia, &entities.CommandMedia{
				ID:        m.ID,
				Name:      m.Name,
				Type:      m.Type,
				URL:       *u,
				Size:      uint64(m.Size),
				CommandID: m.CommandID,
				UserID:    m.UserID,
				CreatedAt: m.CreatedAt,
				UpdatedAt: m.UpdatedAt,
			})
		}

		for _, m := range c.Result.Medias {
			u, err := url.Parse(m.URL)
			if err != nil {
				return nil, fmt.Errorf("%s: %w", op, err)
			}

			cAgg.ResultMedia = append(cAgg.ResultMedia, &entities.ResultMedia{
				ID:        m.ID,
				Name:      m.Name,
				Type:      m.Type,
				URL:       u,
				Size:      int64(m.Size),
				ResultID:  m.ResultID,
				UserID:    m.UserID,
				CreatedAt: m.CreatedAt,
				UpdatedAt: m.UpdatedAt,
			})
		}
	}

	chatAgg, err := aggregates.NewChat(chatEntity, couples)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return chatAgg, nil
}

func (d *Database) SaveResult(ctx context.Context, res *aggregates.Result, logged bool) error {
	const op = "infrastructure.sql.postgres.chatting.SaveResult"

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

	var q string

	if logged {
		q = `insert into 
results (id, text, command_id, tool_id, chat_id, user_id, model_id, created_at, updated_at) 
values ($1, $2, $3, $4, $5, $6, $7, $8, $9)`
	} else {
		q = `insert into 
results (id, text, command_id, tool_id, chat_id, unlogged_user_id, model_id, created_at, updated_at) 
values ($1, $2, $3, $4, $5, $6, $7, $8, $9)`
	}

	_, err = tx.Exec(
		ctx,
		q,
		res.ID,
		res.Text,
		res.CommandID,
		res.ToolID,
		res.ChatID,
		res.UserID,
		res.ModelID,
		res.CreatedAt,
		res.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if logged {
		q = `insert into result_medias 
(id, name, type, url, size, result_id, user_id, created_at, updated_at) 
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`
	} else {
		q = `insert into result_medias 
(id, name, type, url, size, result_id, unlogged_user_id, created_at, updated_at) 
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`
	}

	bh := pgx.Batch{}
	for _, m := range res.Medias {
		bh.Queue(
			q,
			m.ID,
			m.Name,
			m.Type,
			m.URL,
			m.Size,
			m.ResultID,
			m.UserID,
			m.CreatedAt,
			m.UpdatedAt,
		)
	}

	br := tx.SendBatch(ctx, &bh)
	defer br.Close()

	for range res.Medias {
		_, err := br.Exec()
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
	}

	return nil
}

func (d *Database) GetUnloggedUserByIP(
	ctx context.Context,
	ip string,
) (*aggregates.UnloggedUser, error) {
	const op = "infrastructure.sql.postgres.chatting.GetUnloggedUserByIP"

	q := `select 
uu.id as unlogged_user_id, 
uu.ip as unlogged_user_ip,
uu.created_at as unlogged_user_created_at, 
uu.updated_at as unlogged_user_updated_at,
p.id as plan_id,
p.modalities_quote as plan_modalities_quote,
p.subscription_id as plan_subscription_id,
p.created_at as plan_created_at,
p.updated_at as plan_updated_at
from unlogged_users as uu left join plans as p on p.id = uu.plan_id where ip = $1`

	rows, err := d.pool.Query(ctx, q, ip)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	type dto struct {
		models.UnloggedUser
		models.Plan
	}

	dtos, err := pgx.CollectRows(rows, pgx.RowToStructByName[dto])
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if len(dtos) == 0 {
		return nil, domain.NewNotFound(nil, "unlogged user not found", "ip", ip)
	}

	u := dtos[0].UnloggedUser
	p := dtos[0].Plan

	qoute := entities.ModalitiesQuote{}

	if p.ModalitiesQuote.MediaQuote.Valid {
		var quote uint = 0
		if p.ModalitiesQuote.MediaQuote.Int64 < 0 {
			qoute.MediaQuote = &quote
		} else {
			quote = uint(p.ModalitiesQuote.MediaQuote.Int64)
		}
		qoute.MediaQuote = &quote
	}

	if p.ModalitiesQuote.TextContentQuote.Valid {
		var quote uint = 0
		if p.ModalitiesQuote.TextContentQuote.Int64 < 0 {
			qoute.MediaQuote = &quote
		} else {
			quote = uint(p.ModalitiesQuote.TextContentQuote.Int64)
		}
		qoute.MediaQuote = &quote
	}

	if p.ModalitiesQuote.ChattingQuote.Valid {
		var quote uint = 0
		if p.ModalitiesQuote.ChattingQuote.Int64 < 0 {
			qoute.MediaQuote = &quote
		} else {
			quote = uint(p.ModalitiesQuote.ChattingQuote.Int64)
		}
		qoute.MediaQuote = &quote
	}

	agg := &aggregates.UnloggedUser{
		UnloggedUser: &entities.UnloggedUser{
			ID:     u.ID,
			IP:     u.IP,
			PlanID: u.PlanID,
		},
		Plan: &entities.Plan{
			ID:              u.PlanID,
			ModalitiesQuote: qoute,
			SubscriptionID:  p.SubscriptionID,
			CreatedAt:       p.CreatedAt,
			UpdatedAt:       p.UpdatedAt,
		},
	}

	return agg, nil
}

func (d *Database) DeleteChatByID(ctx context.Context, chatID uuid.UUID) error {
	const op = "infrastructure.sql.postgres.chatting.DeleteChatByID"

	userID, ok := ctx.Value(models2.UserIDCtxKey{}).(uuid.UUID)
	if !ok {
		return domain.NewNotFound(nil, "chats not found", "id", chatID.String())
	}

	tx, err := d.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	defer func() {
		if err != nil {
			txErr := tx.Rollback(ctx)
			if txErr != nil {
				err = fmt.Errorf("tx err: %v, rollback err: %w", err, txErr)
				return
			}
		}

		txErr := tx.Commit(ctx)
		if txErr != nil {
			err = fmt.Errorf("commit err: %w", txErr)
		}
	}()

	ct, err := tx.Exec(ctx, `delete from chats where id = $1 and user_id = $2`, chatID, userID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	if ct.RowsAffected() == 0 {
		return domain.NewNotFound(nil, "chats not found", "id", chatID.String())
	}

	return nil
}

func (d *Database) SaveChat(ctx context.Context, chat *aggregates.Chat, logged bool) error {
	const op = "infrastructure.sql.postgres.chatting.SaveChat"

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

	var q string
	if logged {
		q = `insert into chats (id, title, user_id, created_at, updated_at) values ($1, $2, $3, $4, $5)`
		_, err = tx.Exec(ctx, q, chat.ID, chat.Title, chat.UserID, chat.CreatedAt, chat.UpdatedAt)
	} else {
		q = `insert into chats (id, title, unlogged_user_id, created_at, updated_at) values ($1, $2, $3, $4, $5)`
		_, err = tx.Exec(ctx, q, chat.ID, chat.Title, chat.UserID, chat.CreatedAt, chat.UpdatedAt)
	}

	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (d *Database) GetChatByUserID(
	ctx context.Context,
	userID uuid.UUID,
) ([]*aggregates.Chat, error) {
	const op = "infrastructure.sql.postgres.chatting.GetChatByUserID"

	q := `select 
id as chat_id, 
title as chat_title,
created_at as chat_created_at,
updated_at as chat_updated_at,
user_id as user_id,
unlogged_user_id as chat_unlogged_user_id
from chats where user_id = $1 order by updated_at desc`

	rows, err := d.pool.Query(ctx, q, userID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	type chatDTO struct {
		models.Chat
	}

	dtos, err := pgx.CollectRows(rows, pgx.RowToStructByName[chatDTO])
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	result := make([]*aggregates.Chat, 0, len(dtos))
	for _, dto := range dtos {
		var uid uuid.UUID
		if dto.Chat.UserID != uuid.Nil {
			uid = dto.Chat.UserID
		}
		if dto.Chat.UnloggedUserID != uuid.Nil {
			uid = dto.Chat.UnloggedUserID
		}

		chatEntity := &entities.Chat{
			ID:        dto.Chat.ID,
			Title:     dto.Chat.Title,
			CreatedAt: dto.Chat.CreatedAt,
			UpdatedAt: dto.Chat.UpdatedAt,
			UserID:    uid,
		}

		// create aggregate with empty couples
		chatAgg, err := aggregates.NewChat(chatEntity, nil)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}

		result = append(result, chatAgg)
	}

	return result, nil
}

// TODO: remove
func (d *Database) GetMediaByUrls(
	ctx context.Context,
	urls []*url.URL,
) ([]*entities.CommandMedia, error) {
	return nil, nil
}

func (d *Database) saveNewCommand(ctx context.Context, tx pgx.Tx, cm *aggregates.Command) error {
	const op = "infrastructure.sql.postgres.chatting.saveNewCommand"

	q := `insert into commands 
(id, prompt, settings, flags, chat_id, tool_id, model_id, status, user_id, unlogged_user_id, created_at, updated_at) 
values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`

	var toolID *uuid.UUID
	if cm.Tool != nil {
		toolID = &cm.Tool.ID
	}

	_, err := tx.Exec(
		ctx,
		q,
		cm.ID,
		cm.Prompt,
		cm.Settings,
		cm.Flags,
		cm.ChatID,
		toolID,
		cm.ModelID,
		cm.Status,
		cm.UserID,
		nil,
		cm.CreatedAt,
		cm.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
