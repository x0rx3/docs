package postgres

import (
	"context"
	"docs/internal/model"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Session struct {
	pool *pgxpool.Pool
}

func NewSession(pool *pgxpool.Pool) *Session {
	return &Session{
		pool: pool,
	}
}

func (inst *Session) GetSessionByUUID(ctx context.Context, uuid string) (*model.Session, error) {
	session := &model.Session{}
	sql := `SELECT uuid, user_uuid, user_login, expires_at FROM sessions WHERE uuid = $1`

	if err := inst.pool.QueryRow(ctx, sql, uuid).Scan(
		&session.UUID,
		&session.UserUUID,
		&session.UserLogin,
		&session.ExpiresAt,
	); err != nil {
		return nil, err
	}

	return session, nil
}

func (inst *Session) CreateSession(ctx context.Context, session *model.Session) error {
	sql := `INSERT INTO sessions (uuid, user_uuid, user_login, expires_at) VALUEs ($1, $2, $3, $4)`

	if _, err := inst.pool.Exec(ctx, sql, session.UUID, session.UserUUID, session.UserLogin, session.ExpiresAt); err != nil {
		return err
	}

	return nil
}

func (inst *Session) DeleteSession(ctx context.Context, uuid string) error {
	sql := `DELETE FROM session WHERE uuid = $1`

	if _, err := inst.pool.Exec(ctx, sql, uuid); err != nil {
		return err
	}

	return nil
}
