package models

import "github.com/tizzhh/micro-banking/internal/domain/auth/models"

type Currency struct {
	ID   uint64
	Code string
}

type UserWallet struct {
	ID         uint64
	UserID     uint64
	User       models.User
	CurrencyID uint64
	Currency   Currency
	Balance    uint64
}
