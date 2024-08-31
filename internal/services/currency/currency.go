package currency

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	authModels "github.com/tizzhh/micro-banking/internal/domain/auth/models"
	currencyModels "github.com/tizzhh/micro-banking/internal/domain/currency/models"
	currency "github.com/tizzhh/micro-banking/internal/services/currency/errors"
	"github.com/tizzhh/micro-banking/internal/storage"
	"github.com/tizzhh/micro-banking/pkg/logger/sl"
)

func New(log *slog.Logger, currencyOperator CurrencyOperator, userProvider UserProvider, ratesOperator RatesOperator, ratesQuerier RatesQuerier) *Currency {
	return &Currency{
		log:              log,
		currencyOperator: currencyOperator,
		userProvider:     userProvider,
		ratesOperator:    ratesOperator,
		ratesQuerier:     ratesQuerier,
	}
}

type Currency struct {
	log              *slog.Logger
	currencyOperator CurrencyOperator
	userProvider     UserProvider
	ratesOperator    RatesOperator
	ratesQuerier     RatesQuerier
}

type CurrencyOperator interface {
	Buy(ctx context.Context, user authModels.User, currencyCode string, newUserBalance, newCurrencyBalance uint64) error
	Sell(ctx context.Context, user authModels.User, currencyCode string, newUserBalance, newCurrencyBalance uint64) error
	CurrencyBalance(ctx context.Context, user authModels.User, currencyCode string) (uint64, error)
	Wallets(ctx context.Context, user authModels.User) ([]currencyModels.UserWallet, error)
}

type RatesOperator interface {
	SetCurrencyRate(ctx context.Context, currencyCode string, rate float32) error
	GetCurrencyRate(ctx context.Context, currencyCode string) (float32, error)
}

type RatesQuerier interface {
	QueryRates(ctx context.Context, currencyCode string) (float32, error)
}

type UserProvider interface {
	User(ctx context.Context, email string) (authModels.User, error)
}

const (
	priceToCentsConversion = 100
)

func userHasEnoughMoneyToPerformOperation(balance, totalCost uint64) bool {
	return balance >= totalCost
}

func performOperation(balance, currencyBalance, totalCost uint64, isBuy bool) (uint64, uint64) {
	if isBuy {
		return balance - totalCost, currencyBalance + totalCost
	}
	return balance + totalCost, currencyBalance - totalCost
}

func (c *Currency) getUser(ctx context.Context, email string) (authModels.User, error) {
	const caller = "services.currency.getUser"

	log := sl.AddCaller(c.log, caller)

	log.Info("getting user")

	user, err := c.userProvider.User(ctx, email)

	log.Info("user found")

	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			log.Warn("user not found", sl.Error(err))
			return authModels.User{}, fmt.Errorf("%s: %w", caller, currency.ErrUserNotFound)
		}
		log.Error("failed to get user", sl.Error(err))
		return authModels.User{}, fmt.Errorf("%s: %w", caller, err)
	}

	return user, nil
}

func (c *Currency) Buy(ctx context.Context, email string, currencyCode string, amount uint64) (float32, error) {
	const caller = "services.currency.Buy"

	log := sl.AddCaller(c.log, caller)

	log.Info("buying currency")

	user, err := c.getUser(ctx, email)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", caller, err)
	}

	log.Info("user found")

	currencyPrice, err := c.getCurrencyRate(ctx, currencyCode)
	if err != nil {
		log.Error("could not get currency rate", sl.Error(err))
		return 0, fmt.Errorf("%s: %w", caller, err)
	}

	totalCost := uint64(currencyPrice*priceToCentsConversion) * amount

	if !userHasEnoughMoneyToPerformOperation(user.Balance, totalCost) {
		log.Info("not enough money on balance", sl.Error(currency.ErrNotEnoughMoney))
		return 0, fmt.Errorf("%s: %w", caller, currency.ErrNotEnoughMoney)
	}

	log.Info("getting currency balance")

	currencyBalance, err := c.currencyOperator.CurrencyBalance(ctx, user, currencyCode)
	if err != nil {
		log.Error("failed to get currency balance", sl.Error(err))
		return 0, fmt.Errorf("%s: %w", caller, err)
	}

	log.Info("saving balance")

	newBalance, newCurrencyBalance := performOperation(user.Balance, currencyBalance, totalCost, true)
	if err := c.currencyOperator.Buy(ctx, user, currencyCode, newBalance, newCurrencyBalance); err != nil {
		if errors.Is(err, storage.ErrCurrencyCodeNotFound) {
			log.Warn("currency code not found")
			return 0, fmt.Errorf("%s: %w", caller, currency.ErrCurrencyCodeNotFound)
		}
		if errors.Is(err, storage.ErrWalletNotFound) {
			log.Warn("wallet not found")
			return 0, fmt.Errorf("%s: %w", caller, currency.ErrWalletNotFound)
		}
		log.Error("failed to update wallet and user balance", sl.Error(err))
		return 0, fmt.Errorf("%s: %w", caller, err)
	}

	log.Info("currency bought")

	currencyBought := float32((newCurrencyBalance - currencyBalance) / priceToCentsConversion)

	return currencyBought, nil
}

func (c *Currency) Sell(ctx context.Context, email string, currencyCode string, amount uint64) (float32, error) {
	const caller = "services.currency.Sell"

	log := sl.AddCaller(c.log, caller)

	log.Info("selling currency")

	log.Info("user found")

	user, err := c.getUser(ctx, email)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", caller, err)
	}

	currencyPrice, err := c.getCurrencyRate(ctx, currencyCode)
	if err != nil {
		log.Error("could not get currency rate", sl.Error(err))
		return 0, fmt.Errorf("%s: %w", caller, err)
	}

	totalCost := uint64(currencyPrice*priceToCentsConversion) * amount

	log.Info("getting currency balance")

	currencyBalance, err := c.currencyOperator.CurrencyBalance(ctx, user, currencyCode)
	if err != nil {
		log.Error("failed to get currency balance", sl.Error(err))
		return 0, fmt.Errorf("%s: %w", caller, err)
	}

	if !userHasEnoughMoneyToPerformOperation(currencyBalance, totalCost) {
		log.Info("not enough money of currency to sell", sl.Error(currency.ErrNotEnoughCurrency))
		return 0, fmt.Errorf("%s: %w", caller, currency.ErrNotEnoughCurrency)
	}

	log.Info("saving balance")

	newBalance, newCurrencyBalance := performOperation(user.Balance, currencyBalance, totalCost, false)
	if err := c.currencyOperator.Sell(ctx, user, currencyCode, newBalance, newCurrencyBalance); err != nil {
		if errors.Is(err, storage.ErrCurrencyCodeNotFound) {
			log.Warn("currency code not found")
			return 0, fmt.Errorf("%s: %w", caller, currency.ErrCurrencyCodeNotFound)
		}
		if errors.Is(err, storage.ErrWalletNotFound) {
			log.Warn("wallet not found")
			return 0, fmt.Errorf("%s: %w", caller, currency.ErrWalletNotFound)
		}
		log.Error("failed to update wallet and user balance", sl.Error(err))
		return 0, fmt.Errorf("%s: %w", caller, err)
	}

	log.Info("currency sold")

	currencySold := float32((currencyBalance - newCurrencyBalance) / priceToCentsConversion)

	return currencySold, nil
}

func (c *Currency) Wallets(ctx context.Context, email string) ([]currencyModels.UserWallet, error) {
	const caller = "services.currency.Wallets"

	log := sl.AddCaller(c.log, caller)

	log.Info("getting wallet")

	log.Info("user found")

	user, err := c.getUser(ctx, email)
	if err != nil {
		if errors.Is(err, storage.ErrWalletNotFound) {
			log.Warn("wallet not found")
			return nil, fmt.Errorf("%s: %w", caller, currency.ErrWalletNotFound)
		}
		return nil, fmt.Errorf("%s: %w", caller, err)
	}

	wallets, err := c.currencyOperator.Wallets(ctx, user)
	if err != nil {
		log.Error("failed to get user wallets", sl.Error(err))
		return nil, fmt.Errorf("%s: %w", caller, err)
	}

	return wallets, nil
}

func (c *Currency) getCurrencyRate(ctx context.Context, currencyCode string) (float32, error) {
	const caller = "services.currency.getCurrencyRate"

	log := sl.AddCaller(c.log, caller)

	currencyPrice, err := c.ratesOperator.GetCurrencyRate(ctx, currencyCode)
	if errors.Is(err, storage.ErrCurrencyKeyNotFound) {
		log.Warn("currency key not found", slog.String("currency", currencyCode))
		currencyPrice, err = c.requestRatesAPI(ctx, currencyCode)
		if err != nil {
			log.Error("could not request rates", sl.Error(err))
			return 0, fmt.Errorf("%s: %w", caller, err)
		}
		err := c.ratesOperator.SetCurrencyRate(ctx, currencyCode, currencyPrice)
		if err != nil {
			log.Error("could not save currency rates", sl.Error(err))
			return 0, fmt.Errorf("%s: %w", caller, err)
		}
	}
	if err != nil {
		log.Error("internal error", sl.Error(err))
		return 0, fmt.Errorf("%s: %w", caller, err)
	}

	return currencyPrice, nil
}

func (c *Currency) requestRatesAPI(ctx context.Context, currencyCode string) (float32, error) {
	const caller = "services.currency.getCurrencyRate"

	log := sl.AddCaller(c.log, caller)

	rates, err := c.ratesQuerier.QueryRates(ctx, currencyCode)
	if err != nil {
		log.Error("failed to query rates", sl.Error(err))
		return 0, fmt.Errorf("%s: %w", caller, err)
	}
	return rates, nil
}
