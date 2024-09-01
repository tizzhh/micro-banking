package tests

import (
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	currencyv1 "github.com/tizzhh/micro-banking/gen/go/protos/proto/currency"
	suite "github.com/tizzhh/micro-banking/tests/suite/currency"
)

const (
	testUserEmail            = "test@gmail.com"
	testUserPassword         = "admin"
	testCurrencyCode         = "EUR"
	testAmountBuy            = 2
	boughtEuros      float32 = 1.8

	testAmountSell         = 1
	soldEuros      float32 = 0.9

	currencyCodeLen = 3
)

func TestBuySellWallets_HappyPath(t *testing.T) {
	ctx, st := suite.New(t)

	respBuy, err := st.CurrencyClient.Buy(ctx, &currencyv1.BuyRequest{
		Email:        testUserEmail,
		CurrencyCode: testCurrencyCode,
		Amount:       testAmountBuy,
	})

	require.NoError(t, err)
	assert.Equal(t, testUserEmail, respBuy.GetEmail())
	assert.Equal(t, respBuy.GetBought(), boughtEuros)

	respSell, err := st.CurrencyClient.Sell(ctx, &currencyv1.SellRequest{
		Email:        testUserEmail,
		CurrencyCode: testCurrencyCode,
		Amount:       testAmountSell,
	})

	require.NoError(t, err)
	assert.Equal(t, testUserEmail, respSell.GetEmail())
	assert.Equal(t, respSell.GetSold(), soldEuros)

	respWallet, err := st.CurrencyClient.Wallets(ctx, &currencyv1.WalletRequest{
		Email: testUserEmail,
	})

	require.NoError(t, err)
	assert.NotEmpty(t, respWallet)
}

func TestBuy_FailCases(t *testing.T) {
	ctx, st := suite.New(t)

	tests := []struct {
		name         string
		email        string
		currencyCode string
		amount       uint64
		expectedErr  string
	}{
		{
			name:         "Buy with empty email",
			email:        "",
			currencyCode: testCurrencyCode,
			amount:       testAmountBuy,
			expectedErr:  "email: value is empty, which is not a valid email address",
		},
		{
			name:         "Buy with empty currency code",
			email:        gofakeit.Email(),
			currencyCode: "",
			amount:       testAmountBuy,
			expectedErr:  "value must be in list",
		},
		{
			name:         "Buy with empty amount",
			email:        testUserEmail,
			currencyCode: testCurrencyCode,
			amount:       0,
			expectedErr:  "amount: value must be greater than 0",
		},
		{
			name:         "Buy with fake email",
			email:        gofakeit.Email(),
			currencyCode: testCurrencyCode,
			amount:       testAmountBuy,
			expectedErr:  "user not found",
		},
		{
			name:         "Buy with fake currency code",
			email:        gofakeit.Email(),
			currencyCode: gofakeit.LetterN(currencyCodeLen),
			amount:       testAmountBuy,
			expectedErr:  "value must be in list",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := st.CurrencyClient.Buy(ctx, &currencyv1.BuyRequest{
				Email:        tt.email,
				CurrencyCode: tt.currencyCode,
				Amount:       tt.amount,
			})
			require.Error(t, err)
			require.ErrorContains(t, err, tt.expectedErr)
		})
	}
}

func TestSell_FailCases(t *testing.T) {
	ctx, st := suite.New(t)

	tests := []struct {
		name         string
		email        string
		currencyCode string
		amount       uint64
		expectedErr  string
	}{
		{
			name:         "Sell with empty email",
			email:        "",
			currencyCode: testCurrencyCode,
			amount:       testAmountBuy,
			expectedErr:  "email: value is empty, which is not a valid email address",
		},
		{
			name:         "Sell with empty currency code",
			email:        gofakeit.Email(),
			currencyCode: "",
			amount:       testAmountBuy,
			expectedErr:  "value must be in list",
		},
		{
			name:         "Sell with empty amount",
			email:        testUserEmail,
			currencyCode: testCurrencyCode,
			amount:       0,
			expectedErr:  "amount: value must be greater than 0",
		},
		{
			name:         "Sell with fake email",
			email:        gofakeit.Email(),
			currencyCode: testCurrencyCode,
			amount:       testAmountBuy,
			expectedErr:  "user not found",
		},
		{
			name:         "Sell with fake currency code",
			email:        gofakeit.Email(),
			currencyCode: gofakeit.LetterN(currencyCodeLen),
			amount:       testAmountBuy,
			expectedErr:  "value must be in list",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := st.CurrencyClient.Sell(ctx, &currencyv1.SellRequest{
				Email:        tt.email,
				CurrencyCode: tt.currencyCode,
				Amount:       tt.amount,
			})
			require.Error(t, err)
			require.ErrorContains(t, err, tt.expectedErr)
		})
	}
}

func TestWallet_FailCases(t *testing.T) {
	ctx, st := suite.New(t)

	tests := []struct {
		name        string
		email       string
		expectedErr string
	}{
		{
			name:        "Wallets with empty email",
			email:       "",
			expectedErr: "email: value is empty, which is not a valid email address",
		},
		{
			name:        "Wallets with fake email",
			email:       gofakeit.Email(),
			expectedErr: "user not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := st.CurrencyClient.Wallets(ctx, &currencyv1.WalletRequest{
				Email: tt.email,
			})
			require.Error(t, err)
			require.ErrorContains(t, err, tt.expectedErr)
		})
	}
}
