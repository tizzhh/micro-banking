package postgres

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/tizzhh/micro-banking/internal/config"
	"github.com/tizzhh/micro-banking/internal/domain/auth/models"
	"github.com/tizzhh/micro-banking/internal/storage"
)

type Storage struct {
	db *gorm.DB
}

var storageInstance *Storage
var once sync.Once

func Get() (*Storage, error) {
	var err error

	once.Do(func() {
		storageInstance, err = New()
	})

	return storageInstance, err
}

func New() (*Storage, error) {
	const caller = "storage.postgres.New"

	cfg := config.Get()

	db, err := gorm.Open(postgres.Open(fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", cfg.DB.DBHost, cfg.DB.DBPort, cfg.DB.DBUser, cfg.DB.DBPassword, cfg.DB.DBName)))
	if err != nil {
		return nil, fmt.Errorf("%s: %w", caller, err)
	}

	return &Storage{db: db}, nil
}

func (s *Storage) SaveUser(ctx context.Context, user models.User) (int64, error) {
	result := s.db.WithContext(ctx).Create(&user)

	var psqlErr *pgconn.PgError
	if errors.As(result.Error, &psqlErr) && psqlErr.Code == pgerrcode.UniqueViolation {
		return 0, storage.ErrUserAlreadyExists
	}

	return int64(user.ID), result.Error
}

func (s *Storage) User(ctx context.Context, email string) (models.User, error) {
	var user models.User

	result := s.db.WithContext(ctx).First(&user, "email = ?", email)
	if result.RowsAffected == 0 {
		return models.User{}, storage.ErrUserNotFound
	}

	return user, result.Error
}

func (s *Storage) UpdateUser(ctx context.Context, email string, newPassword []byte) error {
	var user models.User

	result := s.db.WithContext(ctx).First(&user, "email = ?", email)
	if result.RowsAffected == 0 {
		return storage.ErrUserNotFound
	}

	user.PassHash = newPassword
	s.db.Save(&user)

	return result.Error
}

func (s *Storage) DeleteUser(ctx context.Context, email string) error {
	var user models.User

	result := s.db.WithContext(ctx).First(&user, "email = ?", email)
	if result.RowsAffected == 0 {
		return storage.ErrUserNotFound
	}

	result = s.db.Delete(&user)

	return result.Error
}
