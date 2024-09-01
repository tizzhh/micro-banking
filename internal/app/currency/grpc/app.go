package grpcapp

import (
	"fmt"
	"log/slog"
	"net"

	currencygrpc "github.com/tizzhh/micro-banking/internal/delivery/grpc/currency"
	"github.com/tizzhh/micro-banking/pkg/logger/sl"
	"google.golang.org/grpc"
)

type App struct {
	log        *slog.Logger
	grpcServer *grpc.Server
	port       int
}

func New(log *slog.Logger, port int, currencyService currencygrpc.Currency, producer currencygrpc.Producer) *App {
	grpcServer := grpc.NewServer()

	currencygrpc.Register(grpcServer, currencyService, producer, log)

	return &App{
		log:        log,
		grpcServer: grpcServer,
		port:       port,
	}
}

func (a *App) MustRun() {
	if err := a.Run(); err != nil {
		panic(err)
	}
}

func (a *App) Run() error {
	const caller = "app.currency.grpc.Run"

	log := sl.AddCaller(a.log, caller)

	l, err := net.Listen("tcp", fmt.Sprintf(":%d", a.port))
	if err != nil {
		return fmt.Errorf("%s: %w", caller, err)
	}

	log.Info("starting grpc server", slog.String("addr", l.Addr().String()))

	if err := a.grpcServer.Serve(l); err != nil {
		return fmt.Errorf("%s: %w", caller, err)
	}

	return nil
}

func (a *App) Stop() {
	const caller = "app.currency.grpc.Stop"

	log := sl.AddCaller(a.log, caller)

	log.Info("stopping grpc server")

	a.grpcServer.GracefulStop()
}
