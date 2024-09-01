package currencyapp

import (
	"context"
	"log/slog"
	"time"

	grpcapp "github.com/tizzhh/micro-banking/internal/app/currency/grpc"
	"github.com/tizzhh/micro-banking/internal/clients/kafka/producer"
	"github.com/tizzhh/micro-banking/internal/services/currency"
	"github.com/tizzhh/micro-banking/internal/storage/postgres"
	"github.com/tizzhh/micro-banking/internal/storage/redis"
	"github.com/tizzhh/micro-banking/pkg/currencyapi"
)

type App struct {
	GRPCServer *grpcapp.App
}

func New(log *slog.Logger, port int, pingTimeout time.Duration, ratesApiTimeout time.Duration, storage *postgres.Storage, producer *producer.Producer) *App {
	cache, err := redis.Get(log)
	if err != nil {
		panic(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), pingTimeout)
	defer cancel()
	if err := cache.MustPing(ctx); err != nil {
		panic(err)
	}

	ratesQuerier := currencyapi.New(log, ratesApiTimeout)

	currencyService := currency.New(log, storage, storage, cache, ratesQuerier)

	grpcApp := grpcapp.New(log, port, currencyService, producer)

	return &App{
		GRPCServer: grpcApp,
	}
}
