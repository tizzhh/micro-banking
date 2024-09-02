package tests

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	currencyApi "github.com/tizzhh/micro-banking/internal/delivery/http/bank/resource/currency"
	currencyMocks "github.com/tizzhh/micro-banking/internal/delivery/http/bank/resource/currency/mocks"
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
