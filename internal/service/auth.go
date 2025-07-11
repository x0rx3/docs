package service

import (
	"context"
	"docs/internal/model"
	"docs/internal/repository"
	"docs/internal/utils"
	"errors"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type Auth struct {
	log         *zap.Logger
	sessionRepo repository.SessionRepository
	userRepo    repository.UserRepository
}

func NewAuth(log *zap.Logger, sessRepo repository.SessionRepository, userRepo repository.UserRepository) *Auth {
	return &Auth{
		log:         log,
		sessionRepo: sessRepo,
		userRepo:    userRepo,
	}
}

func (inst *Auth) Login(ctx context.Context, login, password string) (*model.AuthToken, error) {
	user, err := inst.userRepo.GetUserByLogin(ctx, login)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.ErrorAuthFailed
		}
		return nil, err
	}

	if err := inst.validatePassword(password, user.Password); err != nil {
		return nil, utils.ErrorAuthFailed
	}

	return inst.createSession(ctx, user.UUID, user.Login)
}

func (inst *Auth) Logout(ctx context.Context, token string) error {
	session, err := inst.sessionRepo.GetSessionByUUID(ctx, token)
	if err != nil {
		return utils.ErrorNotFound
	}

	return inst.sessionRepo.DeleteSession(ctx, session.UUID)
}

func (inst *Auth) validatePassword(received, actual string) error {
	return bcrypt.CompareHashAndPassword([]byte(actual), []byte(received))
}

func (inst *Auth) createSession(ctx context.Context, userUUID, userLogin string) (*model.AuthToken, error) {
	sessionUUID := uuid.NewString()

	if err := inst.sessionRepo.CreateSession(ctx, &model.Session{
		UUID:      sessionUUID,
		UserUUID:  userUUID,
		UserLogin: userLogin,
	}); err != nil {
		return nil, err
	}

	return &model.AuthToken{
		AccessToken: sessionUUID,
	}, nil
}
