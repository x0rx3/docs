package service

import (
	"context"
	"docs/internal/model"
	"mime/multipart"
	"time"
)

type AuthService interface {
	Login(ctx context.Context, login, password string) (*model.AuthToken, error)
	Logout(ctx context.Context, token string) error
}

type RegistrationService interface {
	Register(ctx context.Context, token, login, password string) error
}

type DocumentService interface {
	AddDocument(ctx context.Context, document *model.Document, file *multipart.FileHeader) error
	GetDocument(ctx context.Context, uuid, token string) (*model.Document, error)
	ListDocuments(ctx context.Context, token string, data *model.DocumentFilterData) ([]model.Document, error)
	DeleteDocument(ctx context.Context, uuid, token string) error
}

type Cacher interface {
	Get(key string) (any, bool)
	Put(k string, value any, ttl time.Duration, tags []string)
	Invalidate(key string)
	InvalidateByTag(tag string)
	InvalidateByTags(tags []string)
	CleanExpired()
}
