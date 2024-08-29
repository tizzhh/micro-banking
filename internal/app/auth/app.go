package authapp

import (
	"log/slog"
	"time"

	grpcapp "github.com/tizzhh/micro-banking/internal/app/auth/grpc"
	"github.com/tizzhh/micro-banking/internal/services/auth"
	"github.com/tizzhh/micro-banking/internal/storage/postgres"
)

type App struct {
	GRPCServer *grpcapp.App
}

func New(log *slog.Logger, port int, tokenTTL time.Duration) *App {
	storage, err := postgres.Get()

	if err != nil {
		panic(err)
	}

	authService := auth.New(log, tokenTTL, storage, storage, storage, storage)

	grpcApp := grpcapp.New(log, port, tokenTTL, authService)

	return &App{
		GRPCServer: grpcApp,
	}
}
