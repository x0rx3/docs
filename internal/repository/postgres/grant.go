package postgres

import (
	"context"
	"docs/internal/model"
	"docs/internal/utils"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Grant struct {
	pool *pgxpool.Pool
}

func NewGrant(pool *pgxpool.Pool) *Grant {
	return &Grant{
		pool: pool,
	}
}

func (inst *Grant) GetGrantByUserLogin(ctx context.Context, login string) (*model.Grant, error) {
	grant := &model.Grant{}
	sql := `SELECT document_uuid, user_login FROM document_grants WHERE user_login = $1`

	if err := inst.pool.QueryRow(ctx, sql, login).Scan(
		&grant.DocumentUUID,
		&grant.UserLogin,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, utils.ErrorNotFound
		}
	}

	return grant, nil
}

func (inst *Grant) GetGrantByDocumentUUID(ctx context.Context, uuid string) (*model.Grant, error) {
	grant := &model.Grant{}
	sql := `SELECT document_uuid, user_login FROM document_grants WHERE document_uuid = $1`

	if err := inst.pool.QueryRow(ctx, sql, uuid).Scan(
		&grant.DocumentUUID,
		&grant.UserLogin,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, utils.ErrorNotFound
		}
	}

	return grant, nil
}

func (inst *Grant) GetGrantByLoginAndDocUUID(ctx context.Context, uuid, login string) (*model.Grant, error) {
	grant := &model.Grant{}
	sql := `SELECT document_uuid, user_login FROM document_grants WHERE document_uuid = $1 AND user_login = $2`

	if err := inst.pool.QueryRow(ctx, sql, uuid, login).Scan(
		&grant.DocumentUUID,
		&grant.UserLogin,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, utils.ErrorNotFound
		}
		return nil, err
	}

	return grant, nil
}
