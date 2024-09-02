package grpc

import (
	"context"
	"fmt"

	currencyv1 "github.com/tizzhh/micro-banking/gen/go/protos/proto/currency"
	currencyResponse "github.com/tizzhh/micro-banking/internal/delivery/http/bank/resource/currency"
	"github.com/tizzhh/micro-banking/pkg/logger/sl"
)

func (c *Client) Buy(ctx context.Context, email string, currencyCode string, amount uint64) (float32, error) {
	const caller = "clients.currency.grpc.Buy"
	log := sl.AddCaller(c.log, caller)
	log.Info("buying currency")
	resp, err := c.api.Buy(ctx, &currencyv1.BuyRequest{
		Email:        email,
		CurrencyCode: currencyCode,
		Amount:       amount,
	})
	if err != nil {
		log.Error("failed to buy currency", sl.Error(err))
		return 0, fmt.Errorf("%s: %w", caller, err)
	}
	return resp.GetBought(), nil
}

func (c *Client) Sell(ctx context.Context, email string, currencyCode string, amount uint64) (float32, error) {
	const caller = "clients.currency.grpc.Sell"
	log := sl.AddCaller(c.log, caller)
	log.Info("selling currency")
	resp, err := c.api.Sell(ctx, &currencyv1.SellRequest{
		Email:        email,
		CurrencyCode: currencyCode,
		Amount:       amount,
	})
	if err != nil {
		log.Error("failed to sell currency", sl.Error(err))
		return 0, fmt.Errorf("%s: %w", caller, err)
	}
	return resp.GetSold(), nil
}

func (c *Client) Wallets(ctx context.Context, email string) (currencyResponse.WalletResponse, error) {
	const caller = "clients.currency.grpc.Wallets"
	log := sl.AddCaller(c.log, caller)
	log.Info("getting user wallet")
	resp, err := c.api.Wallets(ctx, &currencyv1.WalletRequest{
		Email: email,
	})
	if err != nil {
		log.Error("failed to get user's wallet", sl.Error(err))
		return currencyResponse.WalletResponse{}, fmt.Errorf("%s: %w", caller, err)
	}

	userWalletResp := resp.GetUserWallet()
	userWallet := make([]currencyResponse.Wallet, 0, len(userWalletResp))
	for _, currency := range userWalletResp {
		userWallet = append(userWallet, currencyResponse.Wallet{
			CurrencyCode: currency.GetCurrencyCode(),
			Balance:      currency.GetBalance(),
		})
	}

	return currencyResponse.WalletResponse{Wallets: userWallet}, nil
}
