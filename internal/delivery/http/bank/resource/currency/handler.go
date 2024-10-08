package currency

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
	"github.com/tizzhh/micro-banking/internal/api/validate"
	"github.com/tizzhh/micro-banking/internal/delivery/http/bank/common"
	"github.com/tizzhh/micro-banking/pkg/logger/sl"
)

type CurrencyApi struct {
	log            *slog.Logger
	validator      *validator.Validate
	currencyClient CurrencyClient
}

//go:generate go run github.com/vektra/mockery/v2 --name=CurrencyClient
type CurrencyClient interface {
	Buy(ctx context.Context, email string, currencyCode string, amount uint64) (float32, error)
	Sell(ctx context.Context, email string, currencyCode string, amount uint64) (float32, error)
	Wallets(ctx context.Context, email string) (WalletResponse, error)
}

func New(log *slog.Logger, validator *validator.Validate, currencyClient CurrencyClient) *CurrencyApi {
	return &CurrencyApi{
		log:            log,
		validator:      validator,
		currencyClient: currencyClient,
	}
}

// MyWallet godoc
// @Summary MyWallet
// @Description Return wallet of the user
// @Tags bank
// @Accept json
// @Produce json
// @Param UserRequest body WalletRequest true "Wallet request"
// @Success 200 {object} WalletResponse
// @Failure 400 {object} response.Error
// @Failure 404 {object} response.Error
// @Failure 500 {object} response.Error
// @Router /bank/my-wallet [get]
// @Security BearerAuth
func (ca *CurrencyApi) MyWallet() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const caller = "bank.currency.handler.MyWallet"
		log := sl.AddRequestId(sl.AddCaller(ca.log, caller), middleware.GetReqID(r.Context()))
		log.Info("getting user's wallet")

		var walletRequest WalletRequest

		err := validate.ValidateRequest(ca.log, &walletRequest, r.Body)
		if err != nil {
			log.Error("validation err", sl.Error(err))
			common.HandleValidationErr(w, r, err)
			return
		}

		wallets, err := ca.currencyClient.Wallets(
			r.Context(),
			walletRequest.Email,
		)
		if err != nil {
			log.Error("failed to get user's wallet", sl.Error(err))
			common.HandleGrpcError(ca.log, w, r, err)
			return
		}

		log.Info("user's wallet retieved")

		render.JSON(w, r, wallets)
	}
}

// BuyCurrency godoc
// @Summary Buy currency
// @Description Buy currency
// @Tags currency
// @Accept json
// @Produce json
// @Param BuyRequest body BuyRequest true "Buy request"
// @Success 200 {object} BuyResponse
// @Failure 400 {object} response.Error
// @Failure 404 {object} response.Error
// @Failure 500 {object} response.Error
// @Router /currency/buy [post]
// @Security BearerAuth
func (ca *CurrencyApi) BuyCurrency() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const caller = "bank.currency.handler.Buy"
		log := sl.AddRequestId(sl.AddCaller(ca.log, caller), middleware.GetReqID(r.Context()))
		log.Info("buying currency")

		var buyRequest BuyRequest

		err := validate.ValidateRequest(ca.log, &buyRequest, r.Body)
		if err != nil {
			log.Error("validation err", sl.Error(err))
			common.HandleValidationErr(w, r, err)
			return
		}

		boughtAmount, err := ca.currencyClient.Buy(
			r.Context(),
			buyRequest.Email,
			buyRequest.CurrencyCode,
			buyRequest.Amount,
		)
		if err != nil {
			log.Error("failed to buy currency", sl.Error(err))
			common.HandleGrpcError(ca.log, w, r, err)
			return
		}

		log.Info("currency bought")

		render.JSON(w, r, BuyResponse{
			BoughtAmount: boughtAmount,
			CurrencyCode: buyRequest.CurrencyCode,
		})
	}
}

// SellCurrency godoc
// @Summary Sell currency
// @Description Sell currency
// @Tags currency
// @Accept json
// @Produce json
// @Param SellRequest body SellRequest true "Sell request"
// @Success 200 {object} SellResponse
// @Failure 400 {object} response.Error
// @Failure 404 {object} response.Error
// @Failure 500 {object} response.Error
// @Router /currency/sell [post]
// @Security BearerAuth
func (ca *CurrencyApi) SellCurrency() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const caller = "bank.currency.handler.Sell"
		log := sl.AddRequestId(sl.AddCaller(ca.log, caller), middleware.GetReqID(r.Context()))
		log.Info("selling currency")

		var sellRequest SellRequest

		err := validate.ValidateRequest(ca.log, &sellRequest, r.Body)
		if err != nil {
			log.Error("validation err", sl.Error(err))
			common.HandleValidationErr(w, r, err)
			return
		}

		soldAmount, err := ca.currencyClient.Sell(
			r.Context(),
			sellRequest.Email,
			sellRequest.CurrencyCode,
			sellRequest.Amount,
		)
		if err != nil {
			log.Error("failed to sell currency", sl.Error(err))
			common.HandleGrpcError(ca.log, w, r, err)
			return
		}

		log.Info("currency sold")

		render.JSON(w, r, SellResponse{
			SoldAmount:   soldAmount,
			CurrencyCode: sellRequest.CurrencyCode,
		})
	}
}
