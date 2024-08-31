package auth

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/tizzhh/micro-banking/internal/domain/auth/models"
	"github.com/tizzhh/micro-banking/internal/services/auth/errors"
	"github.com/tizzhh/micro-banking/internal/storage"
	"github.com/tizzhh/micro-banking/pkg/jwt"
	"github.com/tizzhh/micro-banking/pkg/logger/sl"
	"golang.org/x/crypto/bcrypt"
)

func New(log *slog.Logger, tokenTTL time.Duration, userSaver UserSaver, userProvider UserProvider, userUpdater UserUpdater, userDeleter UserDeleter) *Auth {
	return &Auth{
		log:          log,
		tokenTTL:     tokenTTL,
		userSaver:    userSaver,
		userProvider: userProvider,
		userUpdater:  userUpdater,
		userDeleter:  userDeleter,
	}
}

type Auth struct {
	log          *slog.Logger
	tokenTTL     time.Duration
	userSaver    UserSaver
	userProvider UserProvider
	userUpdater  UserUpdater
	userDeleter  UserDeleter
}

type UserSaver interface {
	SaveUser(ctx context.Context, user models.User) (uint64, error)
}

type UserProvider interface {
	User(ctx context.Context, email string) (models.User, error)
}

type UserUpdater interface {
	UpdateUser(ctx context.Context, email string, newPassword []byte) error
}

type UserDeleter interface {
	DeleteUser(ctx context.Context, email string) error
}

func (a *Auth) Register(ctx context.Context, email string, password string, firstName string, lastName string, age uint32) (uint64, error) {
	const caller = "services.auth.Register"

	log := sl.AddCaller(a.log, caller)

	log.Info("registering user")

	passHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Error("failed to generate password hash", sl.Error(err))
		return 0, fmt.Errorf("%s: %w", caller, err)
	}

	newUser := models.User{
		Email:     email,
		PassHash:  passHash,
		FirstName: firstName,
		LastName:  lastName,
		Age:       age,
	}

	newUserId, err := a.userSaver.SaveUser(ctx, newUser)
	if err != nil {
		if errors.Is(err, storage.ErrUserAlreadyExists) {
			log.Warn("user already exists", sl.Error(err))
			return 0, fmt.Errorf("%s: %w", caller, auth.ErrUserAlreadyExists)
		}
		log.Error("failed to save user", sl.Error(err))
		return 0, fmt.Errorf("%s: %w", caller, err)
	}

	log.Info("user saved")

	return newUserId, nil
}

func (a *Auth) Login(ctx context.Context, email string, password string) (string, error) {
	const caller = "services.auth.Login"

	log := sl.AddCaller(a.log, caller)

	log.Info("logging a user in")

	user, err := a.getUserAndCheckPassword(ctx, email, password)
	if err != nil {
		return "", fmt.Errorf("%s: %w", caller, err)
	}

	log.Info("user logged in successfully")

	token, err := jwt.NewToken(user, a.tokenTTL)
	if err != nil {
		log.Error("failed to generate token", sl.Error(err))
		return "", fmt.Errorf("%s: %w", caller, err)
	}

	log.Info("generated token successfully")
	return token, nil
}

func (a *Auth) UpdatePassword(ctx context.Context, email string, oldPassword string, newPassword string) error {
	const caller = "services.auth.UpdatePassword"

	log := sl.AddCaller(a.log, caller)

	log.Info("changing user password")

	_, err := a.getUserAndCheckPassword(ctx, email, oldPassword)
	if err != nil {
		return fmt.Errorf("%s: %w", caller, err)
	}

	log.Info("user retrieved successfully")

	passHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		log.Error("failed to generate password hash", sl.Error(err))
		return fmt.Errorf("%s: %w", caller, err)
	}

	err = a.userUpdater.UpdateUser(ctx, email, passHash)
	if err != nil {
		return fmt.Errorf("%s: %w", caller, err)
	}

	return nil
}

func (a *Auth) Unregister(ctx context.Context, email string, password string) error {
	const caller = "services.auth.Unregister"

	log := sl.AddCaller(a.log, caller)

	log.Info("deleting user")

	_, err := a.getUserAndCheckPassword(ctx, email, password)
	if err != nil {
		return fmt.Errorf("%s: %w", caller, err)
	}

	log.Info("user retrieved successfully")

	err = a.userDeleter.DeleteUser(ctx, email)
	if err != nil {
		return fmt.Errorf("%s: %w", caller, err)
	}

	return nil
}

func (a *Auth) getUserAndCheckPassword(ctx context.Context, email string, password string) (models.User, error) {
	const caller = "services.auth.getUserAndCheckPassword"

	log := sl.AddCaller(a.log, caller)

	log.Info("changing user password")

	user, err := a.userProvider.User(ctx, email)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			log.Warn("user not found", sl.Error(err))
			return models.User{}, fmt.Errorf("%s: %w", caller, auth.ErrInvalidCredentials)
		}
		log.Error("failed to get user", sl.Error(err))
		return models.User{}, fmt.Errorf("%s: %w", caller, err)
	}

	err = bcrypt.CompareHashAndPassword(user.PassHash, []byte(password))
	if err != nil {
		log.Info("invalid credentials", sl.Error(err))
		return models.User{}, fmt.Errorf("%s: %w", caller, auth.ErrInvalidCredentials)
	}

	return user, nil
}
