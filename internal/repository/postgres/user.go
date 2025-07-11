package postgres

import (
	"context"
	"docs/internal/model"
	"docs/internal/utils"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type User struct {
	pool *pgxpool.Pool
}

func NewUser(pool *pgxpool.Pool) *User {
	return &User{
		pool: pool,
	}
}

func (inst *User) GetUserByUUID(ctx context.Context, uuid string) (*model.User, error) {
	var user = &model.User{}
	sql := `SELECT uuid, login, password FROM users WHERE uuid = $1;`
	if err := inst.pool.QueryRow(ctx, sql, uuid).Scan(&user.UUID, &user.Login, &user.Password); err != nil {
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			return nil, utils.ErrorNotFound
		}
		return nil, err
	}
	return user, nil
}

func (inst *User) GetUserByLogin(ctx context.Context, login string) (*model.User, error) {
	var user = &model.User{}
	sql := `SELECT * FROM users WHERE login = $1;`
	if err := inst.pool.QueryRow(ctx, sql, login).Scan(&user.UUID, &user.Login, &user.Password); err != nil {
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			return nil, utils.ErrorNotFound
		}
		return nil, err
	}
	return user, nil
}

func (inst *User) CreateUser(ctx context.Context, user *model.User) error {
	sql := `INSERT INTO users (uuid, login, password) VALUES ($1, $2, $3)`
	_, err := inst.pool.Exec(ctx, sql, user.UUID, user.Login, user.Password)
	if err != nil {
		const errorDublocateKeyCode = "23505"
		if pgerr, ok := err.(*pgconn.PgError); ok && pgerr.Code == errorDublocateKeyCode {
			return utils.ErrorLoginAlradyExists
		}
		return err
	}

	return nil
}
