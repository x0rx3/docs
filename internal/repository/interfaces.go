package repository

import (
	"context"
	"docs/internal/model"
)

type SessionRepository interface {
	GetSessionByUUID(ctx context.Context, uuid string) (*model.Session, error)
	CreateSession(ctx context.Context, session *model.Session) error
	DeleteSession(ctx context.Context, uuid string) error
}

type UserRepository interface {
	GetUserByUUID(ctx context.Context, uuid string) (*model.User, error)
	GetUserByLogin(ctx context.Context, login string) (*model.User, error)
	CreateUser(ctx context.Context, user *model.User) error
}

type DocumentRepository interface {
	CreateDocsWithGrant(ctx context.Context, document *model.Document) error
	GetDocumentWithGrantByUUID(ctx context.Context, uuid string) (*model.Document, error)
	GetDocumentByUUID(ctx context.Context, uuid string) (*model.Document, error)
	ListDocuments(ctx context.Context, data *model.DocumentFilterData) ([]model.Document, error)
	DeleteDocument(ctx context.Context, uuid string) error
}

type GrantRepository interface {
	GetGrantByUserLogin(ctx context.Context, login string) (*model.Grant, error)
	GetGrantByDocumentUUID(ctx context.Context, uuid string) (*model.Grant, error)
	GetGrantByLoginAndDocUUID(ctx context.Context, uuid, login string) (*model.Grant, error)
}
