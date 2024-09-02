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
	currencyApi "github.com/tizzhh/micro-banking/internal/delivery/http/bank/resource/currency"
	currencyMocks "github.com/tizzhh/micro-banking/internal/delivery/http/bank/resource/currency/mocks"
	currency "github.com/tizzhh/micro-banking/internal/services/currency/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	buyRequestTemplate  = `{"amount": %d,"currency_code": "%s","email": "%s"}`
	buyResponseTemplate = `{"bought_amount":%.1f,"currency_code":"%s"}`

	sellRequestTemplate  = `{"amount": %d,"currency_code": "%s","email": "%s"}`
	sellResponseTemplate = `{"sold_amount":%.1f,"currency_code":"%s"}`
)

func TestBuy_HappyPath(t *testing.T) {
	testUserEmail := "test-user0@gmail.com"
	testCurrencyCode := "EUR"
	var testAmount uint64 = 1
	var testBought float32 = 0.9

	reqBody := []byte(fmt.Sprintf(
		buyRequestTemplate,
		testAmount,
		testCurrencyCode,
		testUserEmail,
	))
	bodyReader := bytes.NewBuffer(reqBody)

	req, err := http.NewRequest(http.MethodPost, "/currency/buy", bodyReader)
	require.NoError(t, err)

	mockClient := currencyMocks.NewCurrencyClient(t)
	mockClient.On(
		"Buy",
		context.Background(),
		testUserEmail,
		testCurrencyCode,
		testAmount,
	).Return(testBought, nil)
	currency := currencyApi.New(log, validation, mockClient)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(currency.BuyCurrency())

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, fmt.Sprintf(
		buyResponseTemplate,
		testBought,
		testCurrencyCode,
	), strings.TrimRight(rr.Body.String(), "\n"))
}

func TestBuyHttp_FailCases(t *testing.T) {
	tests := []struct {
		name           string
		email          string
		currencyCode   string
		amount         uint64
		expectedErr    string
		expectedStatus int
	}{
		{
			name:           "Buy with empty email",
			email:          "",
			currencyCode:   testCurrencyCode,
			amount:         testAmountBuy,
			expectedErr:    "field Email is a required field",
			expectedStatus: 400,
		},
		{
			name:           "Buy with empty currency code",
			email:          gofakeit.Email(),
			currencyCode:   "",
			amount:         testAmountBuy,
			expectedErr:    "field CurrencyCode is a required field",
			expectedStatus: 400,
		},
		{
			name:           "Buy with empty amount",
			email:          testUserEmail,
			currencyCode:   testCurrencyCode,
			amount:         0,
			expectedErr:    "field Amount is a required field",
			expectedStatus: 400,
		},
		{
			name:           "Buy with fake currency code",
			email:          gofakeit.Email(),
			currencyCode:   gofakeit.LetterN(currencyCodeLen),
			amount:         testAmountBuy,
			expectedErr:    "field CurrencyCode is not valid",
			expectedStatus: 400,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			reqBody := []byte(fmt.Sprintf(
				buyRequestTemplate,
				tt.amount,
				tt.currencyCode,
				tt.email,
			))
			bodyReader := bytes.NewBuffer(reqBody)

			req, err := http.NewRequest(http.MethodPost, "/currency/buy", bodyReader)
			require.NoError(t, err)

			mockClient := currencyMocks.NewCurrencyClient(t)

			currency := currencyApi.New(log, validation, mockClient)

			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(currency.BuyCurrency())

			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			assert.Equal(t, fmt.Sprintf(
				errorResponseTemplate,
				tt.expectedErr,
			), strings.TrimRight(rr.Body.String(), "\n"))

		})
	}
}

func TestBuyUserDoesNotExist_Fail(t *testing.T) {
	testUserEmail := "test-user0@gmail.com"
	testCurrencyCode := "EUR"
	var testAmount uint64 = 1
	var testBought float32 = 0.9

	expectedError := "not enough money on balance"

	reqBody := []byte(fmt.Sprintf(
		buyRequestTemplate,
		testAmount,
		testCurrencyCode,
		testUserEmail,
	))
	bodyReader := bytes.NewBuffer(reqBody)

	req, err := http.NewRequest(http.MethodPost, "/currency/buy", bodyReader)
	require.NoError(t, err)

	mockClient := currencyMocks.NewCurrencyClient(t)
	mockClient.On(
		"Buy",
		context.Background(),
		testUserEmail,
		testCurrencyCode,
		testAmount,
	).Return(testBought, status.Error(codes.FailedPrecondition, currency.ErrNotEnoughMoney.Error()))
	currency := currencyApi.New(log, validation, mockClient)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(currency.BuyCurrency())

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Equal(t, fmt.Sprintf(
		errorResponseTemplate,
		expectedError,
	), strings.TrimRight(rr.Body.String(), "\n"))
}

func TestBuyNotEnoughMoney_Fail(t *testing.T) {
	testUserEmail := "test-user0@gmail.com"
	testCurrencyCode := "EUR"
	var testAmount uint64 = 1
	var testBought float32 = 0.9

	expectedError := "user not found"

	reqBody := []byte(fmt.Sprintf(
		buyRequestTemplate,
		testAmount,
		testCurrencyCode,
		testUserEmail,
	))
	bodyReader := bytes.NewBuffer(reqBody)

	req, err := http.NewRequest(http.MethodPost, "/currency/buy", bodyReader)
	require.NoError(t, err)

	mockClient := currencyMocks.NewCurrencyClient(t)
	mockClient.On(
		"Buy",
		context.Background(),
		testUserEmail,
		testCurrencyCode,
		testAmount,
	).Return(testBought, status.Error(codes.NotFound, currency.ErrUserNotFound.Error()))
	currency := currencyApi.New(log, validation, mockClient)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(currency.BuyCurrency())

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
	assert.Equal(t, fmt.Sprintf(
		errorResponseTemplate,
		expectedError,
	), strings.TrimRight(rr.Body.String(), "\n"))
}

func TestSell_HappyPath(t *testing.T) {
	testUserEmail := "test-user0@gmail.com"
	testCurrencyCode := "EUR"
	var testAmount uint64 = 1
	var testBought float32 = 0.9

	reqBody := []byte(fmt.Sprintf(
		sellRequestTemplate,
		testAmount,
		testCurrencyCode,
		testUserEmail,
	))
	bodyReader := bytes.NewBuffer(reqBody)

	req, err := http.NewRequest(http.MethodPost, "/currency/sell", bodyReader)
	require.NoError(t, err)

	mockClient := currencyMocks.NewCurrencyClient(t)
	mockClient.On(
		"Sell",
		context.Background(),
		testUserEmail,
		testCurrencyCode,
		testAmount,
	).Return(testBought, nil)
	currency := currencyApi.New(log, validation, mockClient)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(currency.SellCurrency())

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, fmt.Sprintf(
		sellResponseTemplate,
		testBought,
		testCurrencyCode,
	), strings.TrimRight(rr.Body.String(), "\n"))
}

func TestSellHttp_FailCases(t *testing.T) {
	tests := []struct {
		name           string
		email          string
		currencyCode   string
		amount         uint64
		expectedErr    string
		expectedStatus int
	}{
		{
			name:           "Buy with empty email",
			email:          "",
			currencyCode:   testCurrencyCode,
			amount:         testAmountBuy,
			expectedErr:    "field Email is a required field",
			expectedStatus: 400,
		},
		{
			name:           "Buy with empty currency code",
			email:          gofakeit.Email(),
			currencyCode:   "",
			amount:         testAmountBuy,
			expectedErr:    "field CurrencyCode is a required field",
			expectedStatus: 400,
		},
		{
			name:           "Buy with empty amount",
			email:          testUserEmail,
			currencyCode:   testCurrencyCode,
			amount:         0,
			expectedErr:    "field Amount is a required field",
			expectedStatus: 400,
		},
		{
			name:           "Buy with fake currency code",
			email:          gofakeit.Email(),
			currencyCode:   gofakeit.LetterN(currencyCodeLen),
			amount:         testAmountBuy,
			expectedErr:    "field CurrencyCode is not valid",
			expectedStatus: 400,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			reqBody := []byte(fmt.Sprintf(
				buyRequestTemplate,
				tt.amount,
				tt.currencyCode,
				tt.email,
			))
			bodyReader := bytes.NewBuffer(reqBody)

			req, err := http.NewRequest(http.MethodPost, "/currency/sell", bodyReader)
			require.NoError(t, err)

			mockClient := currencyMocks.NewCurrencyClient(t)

			currency := currencyApi.New(log, validation, mockClient)

			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(currency.SellCurrency())

			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			assert.Equal(t, fmt.Sprintf(
				errorResponseTemplate,
				tt.expectedErr,
			), strings.TrimRight(rr.Body.String(), "\n"))

		})
	}
}

func TestSellNotEnoughMoneyOnBalance_Fail(t *testing.T) {
	testUserEmail := "test-user0@gmail.com"
	testCurrencyCode := "EUR"
	var testAmount uint64 = 1
	var testSold float32 = 0.9

	expectedError := "not enough currency on wallet"

	reqBody := []byte(fmt.Sprintf(
		buyRequestTemplate,
		testAmount,
		testCurrencyCode,
		testUserEmail,
	))
	bodyReader := bytes.NewBuffer(reqBody)

	req, err := http.NewRequest(http.MethodPost, "/currency/sell", bodyReader)
	require.NoError(t, err)

	mockClient := currencyMocks.NewCurrencyClient(t)
	mockClient.On(
		"Sell",
		context.Background(),
		testUserEmail,
		testCurrencyCode,
		testAmount,
	).Return(testSold, status.Error(codes.FailedPrecondition, currency.ErrNotEnoughCurrency.Error()))
	currency := currencyApi.New(log, validation, mockClient)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(currency.SellCurrency())

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Equal(t, fmt.Sprintf(
		errorResponseTemplate,
		expectedError,
	), strings.TrimRight(rr.Body.String(), "\n"))
}

func TestSellUserNotFound_Fail(t *testing.T) {
	testUserEmail := "test-user0@gmail.com"
	testCurrencyCode := "EUR"
	var testAmount uint64 = 1
	var testSold float32 = 0.9

	expectedError := "user not found"

	reqBody := []byte(fmt.Sprintf(
		buyRequestTemplate,
		testAmount,
		testCurrencyCode,
		testUserEmail,
	))
	bodyReader := bytes.NewBuffer(reqBody)

	req, err := http.NewRequest(http.MethodPost, "/currency/sell", bodyReader)
	require.NoError(t, err)

	mockClient := currencyMocks.NewCurrencyClient(t)
	mockClient.On(
		"Sell",
		context.Background(),
		testUserEmail,
		testCurrencyCode,
		testAmount,
	).Return(testSold, status.Error(codes.NotFound, currency.ErrUserNotFound.Error()))
	currency := currencyApi.New(log, validation, mockClient)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(currency.SellCurrency())

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
	assert.Equal(t, fmt.Sprintf(
		errorResponseTemplate,
		expectedError,
	), strings.TrimRight(rr.Body.String(), "\n"))
}
