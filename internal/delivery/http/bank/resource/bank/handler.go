package bank

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
	"github.com/tizzhh/micro-banking/internal/api/response"
	"github.com/tizzhh/micro-banking/internal/api/validate"
	"github.com/tizzhh/micro-banking/internal/delivery/http/bank/common"
	bankErrors "github.com/tizzhh/micro-banking/internal/services/bank/errors"
	"github.com/tizzhh/micro-banking/pkg/logger/sl"
)

type BankApi struct {
	log       *slog.Logger
	validator *validator.Validate
	balance   Balancer
}

func New(log *slog.Logger, validator *validator.Validate, balance Balancer) *BankApi {
	return &BankApi{
		log:       log,
		validator: validator,
		balance:   balance,
	}
}

//go:generate go run github.com/vektra/mockery/v2 --name=Balancer
type Balancer interface {
	Deposit(ctx context.Context, email string, amount float32) (float32, error)
	Withdraw(ctx context.Context, email string, amount float32) (float32, error)
}

// Liveness godoc
// @Summary Liveness
// @Description Liveness check
// @Tags bank
// @Produce json
// @Success 200 {object} response.Response
// @Router /liveness [get]
func (ba *BankApi) Liveness() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		response.ReponsdWithOK(w, r, "I'm alive!", http.StatusOK)
	}
}

// Deposit godoc
// @Summary Deposit
// @Description Deposit money to the user's balance account
// @Tags bank
// @Accept json
// @Produce json
// @Param UserRequest body DepositRequest true "Deposit request"
// @Success 200 {object} DepositResponse
// @Failure 400 {object} response.Error
// @Failure 404 {object} response.Error
// @Failure 500 {object} response.Error
// @Router /bank/deposit [post]
// @Security BearerAuth
func (ba *BankApi) Deposit() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const caller = "bank.currency.handler.AddMoneyToBalance"
		log := sl.AddRequestId(sl.AddCaller(ba.log, caller), middleware.GetReqID(r.Context()))
		log.Info("user is making a deposit")

		var depositRequest DepositRequest

		err := validate.ValidateRequest(ba.log, &depositRequest, r.Body)
		if err != nil {
			log.Error("validation err", sl.Error(err))
			common.HandleValidationErr(w, r, err)
			return
		}

		newBalanceAMount, err := ba.balance.Deposit(
			r.Context(),
			depositRequest.Email,
			depositRequest.Amount,
		)
		if err != nil {
			handleBankErr(w, r, err)
			return
		}

		log.Info("deposit completed")

		render.JSON(w, r, DepositResponse{NewBalanceAmount: newBalanceAMount})
	}
}

// Withdraw godoc
// @Summary Deposit
// @Description Withdraw money from the user's balance account
// @Tags bank
// @Accept json
// @Produce json
// @Param UserRequest body WithdrawRequest true "Withdraw request"
// @Success 200 {object} WithdrawResponse
// @Failure 400 {object} response.Error
// @Failure 404 {object} response.Error
// @Failure 500 {object} response.Error
// @Router /bank/withdraw [post]
// @Security BearerAuth
func (ba *BankApi) Withdraw() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const caller = "bank.currency.handler.RemoveMoneyFromBalance"
		log := sl.AddRequestId(sl.AddCaller(ba.log, caller), middleware.GetReqID(r.Context()))
		log.Info("user is making a withdrawal")

		var withdrawRequest WithdrawRequest

		err := validate.ValidateRequest(ba.log, &withdrawRequest, r.Body)
		if err != nil {
			log.Error("validation err", sl.Error(err))
			common.HandleValidationErr(w, r, err)
			return
		}

		newBalanceAMount, err := ba.balance.Withdraw(
			r.Context(),
			withdrawRequest.Email,
			withdrawRequest.Amount,
		)
		if err != nil {
			handleBankErr(w, r, err)
			return
		}

		log.Info("deposit completed")

		render.JSON(w, r, WithdrawResponse{NewBalanceAmount: newBalanceAMount})
	}
}

func handleBankErr(w http.ResponseWriter, r *http.Request, err error) {
	if errors.Is(err, bankErrors.ErrUserNotFound) {
		response.RespondWithError(w, r, bankErrors.ErrUserNotFound.Error(), http.StatusNotFound)
	} else if errors.Is(err, bankErrors.ErrNotEnoughMoney) {
		response.RespondWithError(w, r, bankErrors.ErrNotEnoughMoney.Error(), http.StatusBadRequest)
	} else {
		response.RespondWithError(w, r, "internal error", http.StatusInternalServerError)
	}
}
