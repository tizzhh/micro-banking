package currencyapp

import (
	"log/slog"

	grpcapp "github.com/tizzhh/micro-banking/internal/app/currency/grpc"
	"github.com/tizzhh/micro-banking/internal/services/currency"
	"github.com/tizzhh/micro-banking/internal/storage/postgres"
)

type App struct {
	GRPCServer *grpcapp.App
}

func New(log *slog.Logger, port int) *App {
	storage, err := postgres.Get()

	if err != nil {
		panic(err)
	}

	currencyService := currency.New(log, storage, storage)

	grpcApp := grpcapp.New(log, port, currencyService)

	return &App{
		GRPCServer: grpcApp,
	}
}
