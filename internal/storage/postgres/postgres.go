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
	ctxTx := s.db.WithContext(ctx).Begin()
	defer func() {
		if err := recover(); err != nil {
			ctxTx.Rollback()
		}
	}()

	if err := ctxTx.Error; err != nil {
		return 0, err
	}

	result := ctxTx.Create(&user)

	var psqlErr *pgconn.PgError
	if errors.As(result.Error, &psqlErr) && psqlErr.Code == pgerrcode.UniqueViolation {
		ctxTx.Rollback()
		return 0, storage.ErrUserAlreadyExists
	}
	if result.Error != nil {
		ctxTx.Rollback()
		return 0, result.Error
	}

	if err := InitializeCurrency(ctxTx, user.ID); err != nil {
		ctxTx.Rollback()
		return 0, err
	}

	if err := ctxTx.Commit().Error; err != nil {
		ctxTx.Rollback()
		return 0, err
	}

	return user.ID, nil
}

func (s *Storage) User(ctx context.Context, email string) (authModels.User, error) {
	var user authModels.User

	dbCtx := s.db.WithContext(ctx)

	result := dbCtx.First(&user, "email = ?", email)
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

	ctxTx := s.db.WithContext(ctx).Begin()
	defer func() {
		if err := recover(); err != nil {
			ctxTx.Rollback()
		}
	}()

	if err := ctxTx.Error; err != nil {
		return err
	}

	result := ctxTx.First(&user, "email = ?", email)
	if result.RowsAffected == 0 {
		ctxTx.Rollback()
		return storage.ErrUserNotFound
	}

	if result.Error != nil {
		ctxTx.Rollback()
		return result.Error
	}

	user.PassHash = newPassword
	if result = ctxTx.Save(&user); result.Error != nil {
		ctxTx.Rollback()
		return result.Error
	}

	if err := ctxTx.Commit().Error; err != nil {
		ctxTx.Rollback()
		return err
	}

	return nil
}

func (s *Storage) DeleteUser(ctx context.Context, email string) error {
	var user authModels.User

	ctxTx := s.db.WithContext(ctx).Begin()
	defer func() {
		if err := recover(); err != nil {
			ctxTx.Rollback()
		}
	}()

	if err := ctxTx.Error; err != nil {
		return err
	}

	result := ctxTx.First(&user, "email = ?", email)
	if result.RowsAffected == 0 {
		ctxTx.Rollback()
		return storage.ErrUserNotFound
	}

	if result.Error != nil {
		ctxTx.Rollback()
		return result.Error
	}

	if result = ctxTx.Delete(&user); result.Error != nil {
		ctxTx.Rollback()
		return result.Error
	}

	if err := ctxTx.Commit().Error; err != nil {
		ctxTx.Rollback()
		return err
	}

	return nil
}

func InitializeCurrency(ctxTx *gorm.DB, userId uint64) error {
	var currencies []currencyModels.Currency

	result := ctxTx.Find(&currencies)
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

	result = ctxTx.Create(userWallets)
	if result.Error != nil {
		return result.Error
	}

	return nil
}
