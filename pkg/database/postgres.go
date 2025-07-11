package database

import (
	"context"
	"docs/internal/repository"
	"docs/internal/repository/postgres"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type PostgresRepository struct {
	SessionRepository  repository.SessionRepository
	UserRepository     repository.UserRepository
	DocumentRepository repository.DocumentRepository
	GrantRepository    repository.GrantRepository
}

func NewPostresRepository(log *zap.Logger, dsn string) (*PostgresRepository, error) {
	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		return nil, err
	}
	return &PostgresRepository{
		SessionRepository:  postgres.NewSession(pool),
		UserRepository:     postgres.NewUser(pool),
		DocumentRepository: postgres.NewDocument(log, pool),
		GrantRepository:    postgres.NewGrant(pool),
	}, nil
}
