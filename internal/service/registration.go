package service

import (
	"context"
	"docs/internal/model"
	"docs/internal/repository"
	"docs/internal/utils"
	"fmt"
	"unicode"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

type Registration struct {
	log        *zap.Logger
	adminToken string
	userRepo   repository.UserRepository
}

func NewRegistration(log *zap.Logger, adminToken string, userRepo repository.UserRepository) *Registration {
	return &Registration{
		log:        log,
		adminToken: adminToken,
		userRepo:   userRepo,
	}
}

func (inst *Registration) Register(ctx context.Context, token, login, password string) error {
	if inst.adminToken != token {
		inst.log.Error("unxpected admin token", zap.String("token", token))
		return utils.ErrorInvalidAdminToken
	}

	if err := inst.validateLogin(login); err != nil {
		return err
	}

	if err := inst.validatePassword(password); err != nil {
		return err
	}

	crypPswd, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		inst.log.Error("failed generate password", zap.Error(err))
		return err
	}

	return inst.userRepo.CreateUser(ctx, &model.User{
		UUID:     uuid.NewString(),
		Login:    login,
		Password: string(crypPswd),
	})
}

func (inst *Registration) validateLogin(login string) error {
	for _, r := range login {
		switch {
		case !unicode.IsLetter(r) && !unicode.IsDigit(r):
			inst.log.Error("invalid login for registration, password contains inadmissible symbols")
			return fmt.Errorf("%w: login contains inadmissible symbols: %q", utils.ErrorInvalidLogin, r)
		}
	}

	if len(login) < 8 {
		inst.log.Error("invalid login for registration, inadmissible length of the login")
		return fmt.Errorf("%w: login cannot be less than 8 characters", utils.ErrorInvalidLogin)
	}

	return nil
}

func (inst *Registration) validatePassword(password string) error {
	if len(password) < 8 {
		inst.log.Error("invalid password for registration, inadmissible length of the login")
		return fmt.Errorf("%w: The password should be no less than 8 characters", utils.ErrorInvalidLogin)
	}

	var (
		letters  = 0
		digits   = 0
		specials = 0
	)

	for _, r := range password {
		switch {
		case unicode.IsLetter(r):
			letters++
		case unicode.IsDigit(r):
			digits++
		default:
			specials++
		}
	}

	if letters < 2 {
		inst.log.Error("invalid password, the password does not contain a sufficient amount")
		return fmt.Errorf("%w: The password should contain two letters as a minimum", utils.ErrorInvalidPassword)
	}

	if digits < 1 {
		inst.log.Error("invalid password, the password does not contain a sufficient number of digits")
		return fmt.Errorf("%w: The password should contain at least the bottom the number", utils.ErrorInvalidPassword)
	}

	if specials < 1 {
		inst.log.Error("invalid password, the password does not contain a sufficient number of special characters")
		return fmt.Errorf("%w: The password should be no less than 1 special characters", utils.ErrorInvalidPassword)
	}

	return nil
}
