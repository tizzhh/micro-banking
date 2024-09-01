package bank

import (
	"log/slog"

	httpapp "github.com/tizzhh/micro-banking/internal/app/bank/http"
	authgrpc "github.com/tizzhh/micro-banking/internal/clients/auth/grpc"
	currencygrpc "github.com/tizzhh/micro-banking/internal/clients/currency/grpc"
	"github.com/tizzhh/micro-banking/internal/clients/kafka/producer"
	"github.com/tizzhh/micro-banking/internal/config"
	"github.com/tizzhh/micro-banking/internal/services/bank"
	"github.com/tizzhh/micro-banking/internal/storage/postgres"
)

type App struct {
	HTTPServer *httpapp.App
}

func New(log *slog.Logger, cfg *config.Config, storage *postgres.Storage, producer *producer.Producer) *App {
	authv1Client, err := authgrpc.New(
		log,
		cfg.Clients.AuthClient.Addr,
		cfg.Clients.AuthClient.Timeout,
		cfg.Clients.AuthClient.RetriesCount,
	)
	if err != nil {
		panic(err)
	}

	currencyv1Client, err := currencygrpc.New(
		log,
		cfg.Clients.CurrencyClient.Addr,
		cfg.Clients.CurrencyClient.Timeout,
		cfg.Clients.CurrencyClient.RetriesCount,
	)
	if err != nil {
		panic(err)
	}

	bank := bank.New(log, storage, storage, producer)

	app := httpapp.New(
		log,
		cfg.Http.ReadTimeout,
		cfg.Http.WriteTimeout,
		cfg.Http.IdleTimeout,
		authv1Client,
		currencyv1Client,
		cfg.Http.Port,
		cfg.Http.ShutdownTimeout,
		cfg.TokenTTL,
		bank,
	)
	return &App{HTTPServer: app}
}
