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
	authModels "github.com/tizzhh/micro-banking/internal/domain/auth/models"
	currencyModels "github.com/tizzhh/micro-banking/internal/domain/currency/models"
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

func (s *Storage) SaveUser(ctx context.Context, user authModels.User) (uint64, error) {
	result := s.db.WithContext(ctx).Create(&user)

	var psqlErr *pgconn.PgError
	if errors.As(result.Error, &psqlErr) && psqlErr.Code == pgerrcode.UniqueViolation {
		return 0, storage.ErrUserAlreadyExists
	}
	if result.Error != nil {
		return 0, result.Error
	}

	return user.ID, nil
}

func (s *Storage) User(ctx context.Context, email string) (authModels.User, error) {
	var user authModels.User

	result := s.db.WithContext(ctx).First(&user, "email = ?", email)
	if result.RowsAffected == 0 {
		return authModels.User{}, storage.ErrUserNotFound
	}

	if result.Error != nil {
		return authModels.User{}, result.Error
	}

	return user, nil
}

func (s *Storage) UpdateUser(ctx context.Context, email string, newPassword []byte) error {
	var user authModels.User

	result := s.db.WithContext(ctx).First(&user, "email = ?", email)
	if result.RowsAffected == 0 {
		return storage.ErrUserNotFound
	}

	if result.Error != nil {
		return result.Error
	}

	user.PassHash = newPassword
	result = s.db.WithContext(ctx).Save(&user)
	if result.Error != nil {
		return result.Error
	}

	return nil
}

func (s *Storage) DeleteUser(ctx context.Context, email string) error {
	var user authModels.User

	result := s.db.WithContext(ctx).First(&user, "email = ?", email)
	if result.RowsAffected == 0 {
		return storage.ErrUserNotFound
	}

	if result.Error != nil {
		return result.Error
	}

	result = s.db.WithContext(ctx).Delete(&user)

	if result.Error != nil {
		return result.Error
	}

	return nil
}

func (s *Storage) InitializeCurrency(ctx context.Context, userId uint64) error {
	var currencies []currencyModels.Currency

	result := s.db.WithContext(ctx).Find(&currencies)
	if result.RowsAffected == 0 {
		return storage.ErrUserNotFound
	}

	if result.Error != nil {
		return result.Error
	}

	userWallets := make([]*currencyModels.UserWallet, 0, len(currencies))
	for _, currency := range currencies {
		userWallet := &currencyModels.UserWallet{
			UserID:     userId,
			CurrencyID: currency.ID,
		}
		userWallets = append(userWallets, userWallet)
	}

	result = s.db.WithContext(ctx).Create(userWallets)
	if result.Error != nil {
		return result.Error
	}

	return nil
}
