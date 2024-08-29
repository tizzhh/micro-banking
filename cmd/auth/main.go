package main

import (
	"os"
	"os/signal"
	"syscall"

	authapp "github.com/tizzhh/micro-banking/internal/app/auth"
	"github.com/tizzhh/micro-banking/internal/config"
	"github.com/tizzhh/micro-banking/pkg/logger/sl"
)

func main() {
	const caller = "auth app"

	cfg := config.Get()
	log := sl.AddCaller(sl.Get(), caller)
	log.Info("starting auth app")

	authapp := authapp.New(log, cfg.GRPC.Port, cfg.TokenTTL)
	go authapp.GRPCServer.MustRun()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	<-stop

	authapp.GRPCServer.Stop()
	log.Info("auth app stopped")
}
