package errors

import "errors"

var (
	ErrNotEnoughMoney = errors.New("not enough money on balance")
	ErrUserNotFound   = errors.New("user not found")
)
