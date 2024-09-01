package bank

type DepositRequest struct {
	Email  string  `json:"email" validate:"required,email"`
	Amount float32 `json:"amount" validate:"required,gte=0"`
}

type DepositResponse struct {
	NewBalanceAmount float32 `json:"new_balance_amount" validate:"required,gte=0"`
}

type WithdrawRequest struct {
	Email  string  `json:"email" validate:"required,email"`
	Amount float32 `json:"amount" validate:"required,gte=0"`
}

type WithdrawResponse struct {
	NewBalanceAmount float32 `json:"new_balance_amount" validate:"required,gte=0"`
}
