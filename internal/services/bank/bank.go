package bank

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/tizzhh/micro-banking/internal/domain/auth/models"
	bankErrors "github.com/tizzhh/micro-banking/internal/services/bank/errors"
	"github.com/tizzhh/micro-banking/internal/storage"
	"github.com/tizzhh/micro-banking/pkg/logger/sl"
)

func New(log *slog.Logger, balanceOperator BalanceOperator, userProvider UserProvider, producer Producer) *Bank {
	return &Bank{
		log:             log,
		balanceOperator: balanceOperator,
		userProvider:    userProvider,
		producer:        producer,
	}
}

type Bank struct {
	log             *slog.Logger
	balanceOperator BalanceOperator
	userProvider    UserProvider
	producer        Producer
}

const (
	DepositMsgTemplate    = "Sucessfully made a deposit. New account balance: %f"
	WithdrawalMsgTemplate = "Sucessfully made a withdrawal. New account balance: %f"
)

type Producer interface {
	Produce(emailAddr string, msg string) error
}

type BalanceOperator interface {
	Deposit(ctx context.Context, user models.User, amount uint64) error
	Withdraw(ctx context.Context, user models.User, amount uint64) error
}

type UserProvider interface {
	User(ctx context.Context, email string) (models.User, error)
}

const (
	priceToCentsConversion = 100
)

func (b *Bank) Deposit(ctx context.Context, email string, amount float32) (float32, error) {
	const caller = "services.bank.Deposit"
	log := sl.AddCaller(b.log, caller)
	log.Info("making a deposit")

	amountInUint := uint64(amount * priceToCentsConversion)

	user, err := b.userProvider.User(ctx, email)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			log.Warn("user not found", sl.Error(err))
			return 0, fmt.Errorf("%s: %w", caller, bankErrors.ErrUserNotFound)
		}
		log.Error("failed to get user", sl.Error(err))
		return 0, fmt.Errorf("%s: %w", caller, err)
	}
	newAmount := user.Balance + amountInUint
	err = b.balanceOperator.Deposit(ctx, user, amountInUint)
	if err != nil {
		log.Error("failed to deposit", sl.Error(err))
		return 0, fmt.Errorf("%s: %w", caller, err)
	}

	log.Info("deposit made")
	newBalanceAmount := float32(newAmount / priceToCentsConversion)
	if err = b.producer.Produce(email, fmt.Sprintf(DepositMsgTemplate, newBalanceAmount)); err != nil {
		log.Error("failed to produce", sl.Error(err))
	}

	return newBalanceAmount, nil
}

func (b *Bank) Withdraw(ctx context.Context, email string, amount float32) (float32, error) {
	const caller = "services.bank.Withdraw"
	log := sl.AddCaller(b.log, caller)
	log.Info("making a withdrawal")

	amountInUint := uint64(amount * priceToCentsConversion)

	user, err := b.userProvider.User(ctx, email)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			log.Warn("user not found", sl.Error(err))
			return 0, fmt.Errorf("%s: %w", caller, bankErrors.ErrUserNotFound)
		}
		log.Error("failed to get user", sl.Error(err))
		return 0, fmt.Errorf("%s: %w", caller, err)
	}
	if amountInUint > user.Balance {
		log.Warn("not enough money on balance to withdraw")
		return 0, fmt.Errorf("%s: %w", caller, bankErrors.ErrNotEnoughMoney)
	}

	newAmount := user.Balance - amountInUint
	err = b.balanceOperator.Withdraw(ctx, user, amountInUint)
	if err != nil {
		log.Error("failed to withdraw", sl.Error(err))
		return 0, fmt.Errorf("%s: %w", caller, err)
	}

	log.Info("withdrawal made")
	newBalanceAmount := float32(newAmount / priceToCentsConversion)
	if err = b.producer.Produce(email, fmt.Sprintf(WithdrawalMsgTemplate, newBalanceAmount)); err != nil {
		log.Error("failed to produce", sl.Error(err))
	}

	return newBalanceAmount, nil
}
