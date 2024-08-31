package currency

import "errors"

var (
	ErrNotEnoughMoney       = errors.New("not enough money on balance")
	ErrCurrencyCodeNotFound = errors.New("currency code not found")
	ErrNotEnoughCurrency    = errors.New("not enough currency on wallet")
	ErrUserNotFound         = errors.New("user not found")
)
