package currency

type WalletResponse struct {
	Wallets []Wallet `json:"wallet"`
}

type Wallet struct {
	CurrencyCode string `json:"currency_code"`
	Balance      uint64 `json:"balance"`
}

type WalletRequest struct {
	Email string `json:"email" validate:"required,email"`
}

type BuyRequest struct {
	Email        string `json:"email" validate:"required,email"`
	CurrencyCode string `json:"currency_code" validate:"required,oneof=RUB EUR CNY"`
	Amount       uint64 `json:"amount" validate:"required,gte=0"`
}

type BuyResponse struct {
	BoughtAmount float32 `json:"bought_amount"`
	CurrencyCode string  `json:"currency_code"`
}

type SellRequest struct {
	Email        string `json:"email" validate:"required,email"`
	CurrencyCode string `json:"currency_code" validate:"required,oneof=RUB EUR CNY"`
	Amount       uint64 `json:"amount" validate:"required,gte=0"`
}

type SellResponse struct {
	SoldAmount   float32 `json:"sold_amount"`
	CurrencyCode string  `json:"currency_code"`
}
