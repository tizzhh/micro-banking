package currency

import "errors"

var (
	ErrNotEnoughMoney       = errors.New("not enough money on balance")
	ErrCurrencyCodeNotFound = errors.New("currency code not found")
	ErrNotEnoughCurrency    = errors.New("not enough currency on wallet")
	ErrUserNotFound         = errors.New("user not found")
	ErrWalletNotFound       = errors.New("wallet not found")
	ErrInternal             = errors.New("internal error")
	ErrCurrencyKeyNotFound  = errors.New("currency code not found")
)
