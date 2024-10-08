package tests

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	bankApi "github.com/tizzhh/micro-banking/internal/delivery/http/bank/resource/bank"
	bankMocks "github.com/tizzhh/micro-banking/internal/delivery/http/bank/resource/bank/mocks"
	currencyApi "github.com/tizzhh/micro-banking/internal/delivery/http/bank/resource/currency"
	currencyMocks "github.com/tizzhh/micro-banking/internal/delivery/http/bank/resource/currency/mocks"
	bankErrors "github.com/tizzhh/micro-banking/internal/services/bank/errors"
	currency "github.com/tizzhh/micro-banking/internal/services/currency/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	myWalletRequestTemplate  = `{"email": "%s"}`
	myWalletResponseTemplate = `{"wallet":[{"currency_code":"%s","balance":%d}]}`

	depositRequestTemplate  = `{"amount": %f,"email": "%s"}`
	depositResponseTemplate = `{"new_balance_amount":%.1f}`

	withdrawRequestTemplate  = `{"amount": %f,"email": "%s"}`
	withdrawResponseTemplate = `{"new_balance_amount":%.1f}`
)

func TestHealthCheck_HappyPath(t *testing.T) {
	req, err := http.NewRequest(http.MethodGet, "/liveness", nil)
	require.NoError(t, err)

	bank := bankApi.New(log, validation, bankMocks.NewBalancer(t))

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(bank.Liveness())

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, `{"message":"I'm alive!"}`, strings.TrimRight(rr.Body.String(), "\n"))
}

func TestWallet_HappyPath(t *testing.T) {
	testUserEmail := "test-user0@gmail.com"
	testCurrencyCode := "EUR"
	var testBalance uint64 = 90

	reqBody := []byte(fmt.Sprintf(
		myWalletRequestTemplate,
		testUserEmail,
	))
	bodyReader := bytes.NewBuffer(reqBody)

	req, err := http.NewRequest(http.MethodGet, "/bank/my-wallet", bodyReader)
	require.NoError(t, err)

	expectedResponse := currencyApi.WalletResponse{Wallets: []currencyApi.Wallet{
		{
			CurrencyCode: testCurrencyCode,
			Balance:      testBalance,
		}}}

	mockClient := currencyMocks.NewCurrencyClient(t)
	mockClient.On(
		"Wallets",
		context.Background(),
		testUserEmail,
	).Return(expectedResponse, nil)
	currency := currencyApi.New(log, validation, mockClient)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(currency.MyWallet())

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, fmt.Sprintf(
		myWalletResponseTemplate,
		testCurrencyCode,
		testBalance,
	), strings.TrimRight(rr.Body.String(), "\n"))
}

func TestWalletHttp_FailCases(t *testing.T) {
	tests := []struct {
		name           string
		email          string
		expectedErr    string
		expectedStatus int
	}{
		{
			name:           "Wallets with empty email",
			email:          "",
			expectedErr:    "field Email is a required field",
			expectedStatus: 400,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			reqBody := []byte(fmt.Sprintf(
				myWalletRequestTemplate,
				tt.email,
			))
			bodyReader := bytes.NewBuffer(reqBody)

			req, err := http.NewRequest(http.MethodGet, "/bank/my-wallet", bodyReader)
			require.NoError(t, err)

			mockClient := currencyMocks.NewCurrencyClient(t)
			currency := currencyApi.New(log, validation, mockClient)

			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(currency.MyWallet())

			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			assert.Equal(t, fmt.Sprintf(
				errorResponseTemplate,
				tt.expectedErr,
			), strings.TrimRight(rr.Body.String(), "\n"))

		})
	}
}

func TestWalletUserNotFound_Fail(t *testing.T) {
	testUserEmail := "test-user0@gmail.com"

	expectedErr := "user not found"

	reqBody := []byte(fmt.Sprintf(
		myWalletRequestTemplate,
		testUserEmail,
	))
	bodyReader := bytes.NewBuffer(reqBody)

	req, err := http.NewRequest(http.MethodGet, "/bank/my-wallet", bodyReader)
	require.NoError(t, err)

	mockClient := currencyMocks.NewCurrencyClient(t)
	mockClient.On(
		"Wallets",
		context.Background(),
		testUserEmail,
	).Return(currencyApi.WalletResponse{}, status.Error(codes.NotFound, currency.ErrUserNotFound.Error()))
	currency := currencyApi.New(log, validation, mockClient)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(currency.MyWallet())

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
	assert.Equal(t, fmt.Sprintf(
		errorResponseTemplate,
		expectedErr,
	), strings.TrimRight(rr.Body.String(), "\n"))
}

func TestDeposit_HappyPath(t *testing.T) {
	testUserEmail := "test-user0@gmail.com"
	var testAmount float32 = 1.8

	reqBody := []byte(fmt.Sprintf(
		depositRequestTemplate,
		testAmount,
		testUserEmail,
	))
	bodyReader := bytes.NewBuffer(reqBody)

	req, err := http.NewRequest(http.MethodPost, "/bank/deposit", bodyReader)
	require.NoError(t, err)

	mockClient := bankMocks.NewBalancer(t)
	mockClient.On(
		"Deposit",
		context.Background(),
		testUserEmail,
		testAmount,
	).Return(testAmount, nil)
	bank := bankApi.New(log, validation, mockClient)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(bank.Deposit())

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, fmt.Sprintf(
		depositResponseTemplate,
		testAmount,
	), strings.TrimRight(rr.Body.String(), "\n"))
}

func TestDepositHttp_FailCases(t *testing.T) {
	var testAmount float32 = 1.8

	tests := []struct {
		name           string
		email          string
		amount         float32
		expectedErr    string
		expectedStatus int
	}{
		{
			name:           "Empty email",
			email:          "",
			amount:         testAmount,
			expectedErr:    "field Email is a required field",
			expectedStatus: 400,
		},
		{
			name:           "Empty amount",
			email:          gofakeit.Email(),
			amount:         0,
			expectedErr:    "field Amount is a required field",
			expectedStatus: 400,
		},
		{
			name:           "Invalid email",
			email:          "askdlaskd",
			amount:         testAmount,
			expectedErr:    "field Email is not a valid email",
			expectedStatus: 400,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			reqBody := []byte(fmt.Sprintf(
				depositRequestTemplate,
				tt.amount,
				tt.email,
			))
			bodyReader := bytes.NewBuffer(reqBody)

			req, err := http.NewRequest(http.MethodPost, "/bank/deposit", bodyReader)
			require.NoError(t, err)

			mockClient := bankMocks.NewBalancer(t)
			bank := bankApi.New(log, validation, mockClient)

			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(bank.Deposit())

			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			assert.Equal(t, fmt.Sprintf(
				errorResponseTemplate,
				tt.expectedErr,
			), strings.TrimRight(rr.Body.String(), "\n"))

		})
	}

}

func TestDepositUserNotFound_Fail(t *testing.T) {
	testUserEmail := "test-user0@gmail.com"
	var testAmount float32 = 1.8

	expectedErr := "user not found"

	reqBody := []byte(fmt.Sprintf(
		depositRequestTemplate,
		testAmount,
		testUserEmail,
	))
	bodyReader := bytes.NewBuffer(reqBody)

	req, err := http.NewRequest(http.MethodPost, "/bank/deposit", bodyReader)
	require.NoError(t, err)

	mockClient := bankMocks.NewBalancer(t)
	mockClient.On(
		"Deposit",
		context.Background(),
		testUserEmail,
		testAmount,
	).Return(testAmount, bankErrors.ErrUserNotFound)
	bank := bankApi.New(log, validation, mockClient)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(bank.Deposit())

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
	assert.Equal(t, fmt.Sprintf(
		errorResponseTemplate,
		expectedErr,
	), strings.TrimRight(rr.Body.String(), "\n"))
}

func TestWithdraw_HappyPath(t *testing.T) {
	testUserEmail := "test-user0@gmail.com"
	var testAmount float32 = 1.8

	reqBody := []byte(fmt.Sprintf(
		withdrawRequestTemplate,
		testAmount,
		testUserEmail,
	))
	bodyReader := bytes.NewBuffer(reqBody)

	req, err := http.NewRequest(http.MethodPost, "/bank/withdraw", bodyReader)
	require.NoError(t, err)

	mockClient := bankMocks.NewBalancer(t)
	mockClient.On(
		"Withdraw",
		context.Background(),
		testUserEmail,
		testAmount,
	).Return(testAmount, nil)
	bank := bankApi.New(log, validation, mockClient)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(bank.Withdraw())

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, fmt.Sprintf(
		withdrawResponseTemplate,
		testAmount,
	), strings.TrimRight(rr.Body.String(), "\n"))
}

func TestWithdrawHttp_FailCases(t *testing.T) {
	var testAmount float32 = 1.8

	tests := []struct {
		name           string
		email          string
		amount         float32
		expectedErr    string
		expectedStatus int
	}{
		{
			name:           "Empty email",
			email:          "",
			amount:         testAmount,
			expectedErr:    "field Email is a required field",
			expectedStatus: 400,
		},
		{
			name:           "Empty amount",
			email:          gofakeit.Email(),
			amount:         0,
			expectedErr:    "field Amount is a required field",
			expectedStatus: 400,
		},
		{
			name:           "Invalid email",
			email:          "askdlaskd",
			amount:         testAmount,
			expectedErr:    "field Email is not a valid email",
			expectedStatus: 400,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			reqBody := []byte(fmt.Sprintf(
				withdrawRequestTemplate,
				tt.amount,
				tt.email,
			))
			bodyReader := bytes.NewBuffer(reqBody)

			req, err := http.NewRequest(http.MethodPost, "/bank/withdraw", bodyReader)
			require.NoError(t, err)

			mockClient := bankMocks.NewBalancer(t)
			bank := bankApi.New(log, validation, mockClient)

			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(bank.Withdraw())

			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			assert.Equal(t, fmt.Sprintf(
				errorResponseTemplate,
				tt.expectedErr,
			), strings.TrimRight(rr.Body.String(), "\n"))

		})
	}

}

func TestWithdrawUserNotFound_Fail(t *testing.T) {
	testUserEmail := "test-user0@gmail.com"
	var testAmount float32 = 1.8

	expectedErr := "user not found"

	reqBody := []byte(fmt.Sprintf(
		withdrawRequestTemplate,
		testAmount,
		testUserEmail,
	))
	bodyReader := bytes.NewBuffer(reqBody)

	req, err := http.NewRequest(http.MethodPost, "/bank/withdraw", bodyReader)
	require.NoError(t, err)

	mockClient := bankMocks.NewBalancer(t)
	mockClient.On(
		"Withdraw",
		context.Background(),
		testUserEmail,
		testAmount,
	).Return(testAmount, bankErrors.ErrUserNotFound)
	bank := bankApi.New(log, validation, mockClient)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(bank.Withdraw())

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
	assert.Equal(t, fmt.Sprintf(
		errorResponseTemplate,
		expectedErr,
	), strings.TrimRight(rr.Body.String(), "\n"))
}

func TestWithdrawNotEnoughMoney_Fail(t *testing.T) {
	testUserEmail := "test-user0@gmail.com"
	var testAmount float32 = 1.8

	expectedErr := "not enough money on balance"

	reqBody := []byte(fmt.Sprintf(
		withdrawRequestTemplate,
		testAmount,
		testUserEmail,
	))
	bodyReader := bytes.NewBuffer(reqBody)

	req, err := http.NewRequest(http.MethodPost, "/bank/withdraw", bodyReader)
	require.NoError(t, err)

	mockClient := bankMocks.NewBalancer(t)
	mockClient.On(
		"Withdraw",
		context.Background(),
		testUserEmail,
		testAmount,
	).Return(testAmount, bankErrors.ErrNotEnoughMoney)
	bank := bankApi.New(log, validation, mockClient)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(bank.Withdraw())

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Equal(t, fmt.Sprintf(
		errorResponseTemplate,
		expectedErr,
	), strings.TrimRight(rr.Body.String(), "\n"))
}
