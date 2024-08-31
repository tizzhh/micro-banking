package storage

import "errors"

var (
	ErrUserAlreadyExists    = errors.New("user already exists")
	ErrUserNotFound         = errors.New("user not found")
	ErrCurrencyCodeNotFound = errors.New("currency code not found")
	ErrWalletNotFound       = errors.New("wallet not found")
)
