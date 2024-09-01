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
	balance   Balance
}

func New(log *slog.Logger, validator *validator.Validate, balance Balance) *BankApi {
	return &BankApi{
		log:       log,
		validator: validator,
		balance:   balance,
	}
}

type Balance interface {
	Deposit(ctx context.Context, email string, amount float32) (float32, error)
	Withdraw(ctx context.Context, email string, amount float32) (float32, error)
}

func (ba *BankApi) Liveness() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		response.ReponsdWithOK(w, r, "I'm alive!", http.StatusOK)
	}
}

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

		render.JSON(w, r, DepositResponse{NewBalanceAmount: newBalanceAMount})
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
