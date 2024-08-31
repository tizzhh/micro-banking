package main

import (
	"os"
	"os/signal"
	"syscall"

	currencyapp "github.com/tizzhh/micro-banking/internal/app/currency"
	"github.com/tizzhh/micro-banking/internal/config"
	"github.com/tizzhh/micro-banking/pkg/logger/sl"
)

func main() {
	cfg := config.Get()
	log := sl.Get()
	log.Info("starting auth app")

	currencyApp := currencyapp.New(log, cfg.GRPC.Port)
	go currencyApp.GRPCServer.MustRun()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	<-stop

	currencyApp.GRPCServer.Stop()
	log.Info("auth app stopped")
}
