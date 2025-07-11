package service

import (
	"docs/internal/service"
	"docs/pkg/database"

	"go.uber.org/zap"
)

type ServiceCollector struct {
	AuthService         service.AuthService
	RegistrationService service.RegistrationService
	DocumentService     service.DocumentService
}

func NewServiceCollector(log *zap.Logger, uploadPath, adminToken string, repo *database.PostgresRepository) *ServiceCollector {
	cache := NewInternalCache()
	docsService := service.NewAuth(log, repo.SessionRepository, repo.UserRepository)
	registrationService := service.NewRegistration(log, adminToken, repo.UserRepository)
	documentService := service.NewDocument(log, uploadPath, repo.GrantRepository, repo.DocumentRepository, repo.SessionRepository, cache)

	return &ServiceCollector{
		AuthService:         docsService,
		RegistrationService: registrationService,
		DocumentService:     documentService,
	}
}
