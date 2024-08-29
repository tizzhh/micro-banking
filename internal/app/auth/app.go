package authapp

import (
	"log/slog"
	"time"

	grpcapp "github.com/tizzhh/micro-banking/internal/app/auth/grpc"
)

type App struct {
	GRPCServer *grpcapp.App
}

func New(log *slog.Logger, port int, tokenTTL time.Duration) *App {
	// TODO: get storage

	// TODO: get service

	grpcApp := grpcapp.New(log, port)

	return &App{
		GRPCServer: grpcApp,
	}
}
