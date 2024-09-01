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
	const caller = "storage.postgres.SaveUser"

	ctxTx := s.db.WithContext(ctx).Begin()
	defer func() {
		if err := recover(); err != nil {
			ctxTx.Rollback()
		}
	}()

	if err := ctxTx.Error; err != nil {
		return 0, fmt.Errorf("%s: %w", caller, err)
	}

	result := ctxTx.Create(&user)

	var psqlErr *pgconn.PgError
	if errors.As(result.Error, &psqlErr) && psqlErr.Code == pgerrcode.UniqueViolation {
		ctxTx.Rollback()
		return 0, fmt.Errorf("%s: %w", caller, storage.ErrUserAlreadyExists)
	}
	if result.Error != nil {
		ctxTx.Rollback()
		return 0, fmt.Errorf("%s: %w", caller, result.Error)
	}

	if err := InitializeCurrency(ctxTx, user.ID); err != nil {
		ctxTx.Rollback()
		return 0, fmt.Errorf("%s: %w", caller, err)
	}

	if err := ctxTx.Commit().Error; err != nil {
		ctxTx.Rollback()
		return 0, fmt.Errorf("%s: %w", caller, err)
	}

	return user.ID, nil
}

func (s *Storage) User(ctx context.Context, email string) (authModels.User, error) {
	const caller = "storage.postgres.User"

	var user authModels.User

	dbCtx := s.db.WithContext(ctx)

	result := dbCtx.First(&user, "email = ?", email)
	if result.RowsAffected == 0 {
		return authModels.User{}, fmt.Errorf("%s: %w", caller, storage.ErrUserNotFound)
	}

	if result.Error != nil {
		return authModels.User{}, fmt.Errorf("%s: %w", caller, result.Error)
	}

	return user, nil
}

func (s *Storage) UpdateUser(ctx context.Context, email string, newPassword []byte) error {
	const caller = "storage.postgres.UpdateUser"

	var user authModels.User

	ctxTx := s.db.WithContext(ctx).Begin()
	defer func() {
		if err := recover(); err != nil {
			ctxTx.Rollback()
		}
	}()

	if err := ctxTx.Error; err != nil {
		return fmt.Errorf("%s: %w", caller, err)
	}

	result := ctxTx.First(&user, "email = ?", email)
	if result.RowsAffected == 0 {
		ctxTx.Rollback()
		return fmt.Errorf("%s: %w", caller, storage.ErrUserNotFound)
	}

	if result.Error != nil {
		ctxTx.Rollback()
		return fmt.Errorf("%s: %w", caller, result.Error)
	}

	user.PassHash = newPassword
	if result = ctxTx.Save(&user); result.Error != nil {
		ctxTx.Rollback()
		return fmt.Errorf("%s: %w", caller, result.Error)
	}

	if err := ctxTx.Commit().Error; err != nil {
		ctxTx.Rollback()
		return fmt.Errorf("%s: %w", caller, err)
	}

	return nil
}

func (s *Storage) DeleteUser(ctx context.Context, email string) error {
	const caller = "storage.postgres.DeleteUser"

	var user authModels.User

	ctxTx := s.db.WithContext(ctx).Begin()
	defer func() {
		if err := recover(); err != nil {
			ctxTx.Rollback()
		}
	}()

	if err := ctxTx.Error; err != nil {
		return fmt.Errorf("%s: %w", caller, err)
	}

	result := ctxTx.First(&user, "email = ?", email)
	if result.RowsAffected == 0 {
		ctxTx.Rollback()
		return fmt.Errorf("%s: %w", caller, storage.ErrUserNotFound)
	}

	if result.Error != nil {
		ctxTx.Rollback()
		return fmt.Errorf("%s: %w", caller, result.Error)
	}

	if result = ctxTx.Delete(&user); result.Error != nil {
		ctxTx.Rollback()
		return fmt.Errorf("%s: %w", caller, result.Error)
	}

	if err := ctxTx.Commit().Error; err != nil {
		ctxTx.Rollback()
		return fmt.Errorf("%s: %w", caller, err)
	}

	return nil
}

func InitializeCurrency(ctxTx *gorm.DB, userId uint64) error {
	const caller = "storage.postgres.InitializeCurrency"

	var currencies []currencyModels.Currency

	result := ctxTx.Find(&currencies)
	if result.RowsAffected == 0 {
		return fmt.Errorf("%s: %w", caller, storage.ErrUserNotFound)
	}

	if result.Error != nil {
		return fmt.Errorf("%s: %w", caller, result.Error)
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
		return fmt.Errorf("%s: %w", caller, result.Error)
	}

	return nil
}

func getCurrency(ctxTx *gorm.DB, currencyCode string) (currencyModels.Currency, error) {
	const caller = "storage.postgres.getCurrency"

	var currency currencyModels.Currency

	result := ctxTx.Where(currencyModels.Currency{Code: currencyCode}).First(&currency)
	if result.RowsAffected == 0 {
		return currencyModels.Currency{}, fmt.Errorf("%s: %w", caller, storage.ErrCurrencyCodeNotFound)
	}
	if result.Error != nil {
		return currencyModels.Currency{}, fmt.Errorf("%s: %w", caller, result.Error)
	}

	return currency, nil
}

func getWallet(ctxTx *gorm.DB, user authModels.User, currency currencyModels.Currency) (currencyModels.UserWallet, error) {
	const caller = "storage.postgres.getWallet"

	var wallet currencyModels.UserWallet
	result := ctxTx.Where("user_id = ? AND currency_id = ?", user.ID, currency.ID).First(&wallet)
	if result.RowsAffected == 0 {
		return currencyModels.UserWallet{}, fmt.Errorf("%s: %w", caller, storage.ErrWalletNotFound)
	}
	if result.Error != nil {
		return currencyModels.UserWallet{}, fmt.Errorf("%s: %w", caller, result.Error)
	}

	return wallet, nil
}

func (s *Storage) performBuySellOperation(ctx context.Context, user authModels.User, currencyCode string, newUserBalance, newCurrencyBalance uint64) error {
	const caller = "storage.postgres.performBuySellOperation"

	ctxTx := s.db.WithContext(ctx).Begin()
	defer func() {
		if err := recover(); err != nil {
			ctxTx.Rollback()
		}
	}()

	currency, err := getCurrency(ctxTx, currencyCode)
	if err != nil {
		ctxTx.Rollback()
		return fmt.Errorf("%s: %w", caller, err)
	}

	wallet, err := getWallet(ctxTx, user, currency)
	if err != nil {
		ctxTx.Rollback()
		return fmt.Errorf("%s: %w", caller, err)
	}

	user.Balance = newUserBalance
	if err := ctxTx.Save(&user).Error; err != nil {
		ctxTx.Rollback()
		return fmt.Errorf("%s: %w", caller, err)
	}

	wallet.Balance = newCurrencyBalance
	if err := ctxTx.Save(&wallet).Error; err != nil {
		ctxTx.Rollback()
		return fmt.Errorf("%s: %w", caller, err)
	}

	if err := ctxTx.Commit().Error; err != nil {
		ctxTx.Rollback()
		return fmt.Errorf("%s: %w", caller, err)
	}

	return nil
}

func (s *Storage) Buy(ctx context.Context, user authModels.User, currencyCode string, newUserBalance, newCurrencyBalance uint64) error {
	const caller = "storage.postgres.Buy"

	if err := s.performBuySellOperation(ctx, user, currencyCode, newUserBalance, newCurrencyBalance); err != nil {
		return fmt.Errorf("%s: %w", caller, err)
	}

	return nil
}

func (s *Storage) Sell(ctx context.Context, user authModels.User, currencyCode string, newUserBalance, newCurrencyBalance uint64) error {
	const caller = "storage.postgres.Sell"

	if err := s.performBuySellOperation(ctx, user, currencyCode, newUserBalance, newCurrencyBalance); err != nil {
		return fmt.Errorf("%s: %w", caller, err)
	}

	return nil
}

func (s *Storage) CurrencyBalance(ctx context.Context, user authModels.User, currencyCode string) (uint64, error) {
	const caller = "storage.postgres.CurrencyBalance"

	ctxDb := s.db.WithContext(ctx)
	currency, err := getCurrency(ctxDb, currencyCode)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", caller, err)
	}

	wallet, err := getWallet(ctxDb, user, currency)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", caller, err)
	}

	return wallet.Balance, nil
}

func (s *Storage) Wallets(ctx context.Context, user authModels.User) ([]currencyModels.UserWallet, error) {
	const caller = "storage.postgres.Wallet"

	var userWallets []currencyModels.UserWallet

	ctxDb := s.db.WithContext(ctx)
	result := ctxDb.Preload("Currency").Find(&userWallets)
	if result.RowsAffected == 0 {
		return nil, fmt.Errorf("%s: %w", caller, storage.ErrWalletNotFound)
	}
	if result.Error != nil {
		return nil, fmt.Errorf("%s: %w", caller, result.Error)
	}

	return userWallets, nil
}

func (s *Storage) updateUserBalance(ctx context.Context, user authModels.User, newBalanceAmount uint64) error {
	const caller = "storage.postgres.updateUserBalance"

	ctxTx := s.db.WithContext(ctx).Begin()
	defer func() {
		if err := recover(); err != nil {
			ctxTx.Rollback()
		}
	}()

	user.Balance = newBalanceAmount
	result := ctxTx.Save(&user)
	if result.Error != nil {
		ctxTx.Rollback()
		return fmt.Errorf("%s: %w", caller, result.Error)
	}

	if err := ctxTx.Commit().Error; err != nil {
		ctxTx.Rollback()
		return fmt.Errorf("%s: %w", caller, err)
	}
	return nil
}

func (s *Storage) Deposit(ctx context.Context, user authModels.User, newBalanceAmount uint64) error {
	const caller = "storage.postgres.Deposit"
	if err := s.updateUserBalance(ctx, user, newBalanceAmount); err != nil {
		return fmt.Errorf("%s: %w", caller, err)
	}
	return nil
}

func (s *Storage) Withdraw(ctx context.Context, user authModels.User, newBalanceAmount uint64) error {
	const caller = "storage.postgres.Withdraw"
	if err := s.updateUserBalance(ctx, user, newBalanceAmount); err != nil {
		return fmt.Errorf("%s: %w", caller, err)
	}
	return nil
}

func (s *Storage) Stop() error {
	sqlDB, err := s.db.DB()
	if err != nil {
		return err
	}
	err = sqlDB.Close()
	if err != nil {
		return err
	}
	return nil
}
