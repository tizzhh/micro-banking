package currency

import (
	"context"
	"errors"

	"github.com/bufbuild/protovalidate-go"
	currencyv1 "github.com/tizzhh/micro-banking/gen/go/protos/proto/currency"
	"github.com/tizzhh/micro-banking/internal/domain/currency/models"
	currency "github.com/tizzhh/micro-banking/internal/services/currency/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type serverApi struct {
	currencyv1.UnimplementedCurrencyServer
	currency Currency
}

func Register(gRPC *grpc.Server, currency Currency) {
	currencyv1.RegisterCurrencyServer(gRPC, &serverApi{currency: currency})
}

type Currency interface {
	Buy(ctx context.Context, email string, currencyCode string, amount uint64) (float32, error)
	Sell(ctx context.Context, email string, currencyCode string, amount uint64) (float32, error)
	Wallets(ctx context.Context, email string) ([]models.UserWallet, error)
}

func (s *serverApi) Buy(ctx context.Context, req *currencyv1.BuyRequest) (*currencyv1.BuyResponse, error) {
	validator, err := protovalidate.New()
	if err != nil {
		return nil, status.Error(codes.Internal, currency.ErrInternal.Error())
	}

	if err = validator.Validate(req); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	boughtAmount, err := s.currency.Buy(ctx, req.GetEmail(), req.GetCurrencyCode(), req.GetAmount())
	if err != nil {
		if errors.Is(err, currency.ErrNotEnoughMoney) {
			return nil, status.Error(codes.FailedPrecondition, currency.ErrNotEnoughMoney.Error())
		}
		if errors.Is(err, currency.ErrCurrencyCodeNotFound) {
			return nil, status.Error(codes.NotFound, currency.ErrCurrencyCodeNotFound.Error())
		}
		if errors.Is(err, currency.ErrUserNotFound) {
			return nil, status.Error(codes.NotFound, currency.ErrUserNotFound.Error())
		}
		return nil, status.Error(codes.Internal, currency.ErrInternal.Error())
	}

	return &currencyv1.BuyResponse{Email: req.GetEmail(), Bought: boughtAmount}, nil
}

func (s *serverApi) Sell(ctx context.Context, req *currencyv1.SellRequest) (*currencyv1.SellResponse, error) {
	validator, err := protovalidate.New()
	if err != nil {
		return nil, status.Error(codes.Internal, currency.ErrInternal.Error())
	}

	if err = validator.Validate(req); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	soldAmount, err := s.currency.Sell(ctx, req.GetEmail(), req.GetCurrencyCode(), req.GetAmount())
	if err != nil {
		if errors.Is(err, currency.ErrNotEnoughCurrency) {
			return nil, status.Error(codes.FailedPrecondition, currency.ErrNotEnoughCurrency.Error())
		}
		if errors.Is(err, currency.ErrCurrencyCodeNotFound) {
			return nil, status.Error(codes.NotFound, currency.ErrCurrencyCodeNotFound.Error())
		}
		if errors.Is(err, currency.ErrUserNotFound) {
			return nil, status.Error(codes.NotFound, currency.ErrUserNotFound.Error())
		}
		return nil, status.Error(codes.Internal, currency.ErrInternal.Error())
	}
	return &currencyv1.SellResponse{Email: req.GetEmail(), Sold: soldAmount}, nil
}

func (s *serverApi) Wallets(ctx context.Context, req *currencyv1.WalletRequest) (*currencyv1.WalletResponse, error) {
	validator, err := protovalidate.New()
	if err != nil {
		return nil, status.Error(codes.Internal, currency.ErrInternal.Error())
	}

	if err = validator.Validate(req); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	wallet, err := s.currency.Wallets(ctx, req.GetEmail())
	if err != nil {
		if errors.Is(err, currency.ErrUserNotFound) {
			return nil, status.Error(codes.NotFound, currency.ErrUserNotFound.Error())
		}
		return nil, status.Error(codes.Internal, currency.ErrInternal.Error())
	}

	wallets := make([]*currencyv1.UserWallet, 0, len(wallet))
	for _, currency := range wallet {
		wallets = append(wallets, &currencyv1.UserWallet{CurrencyCode: currency.Currency.Code, Balance: currency.Balance})
	}

	return &currencyv1.WalletResponse{UserWallet: wallets}, nil
}
