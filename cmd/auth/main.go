package main

import (
	"os"
	"os/signal"
	"syscall"

	authapp "github.com/tizzhh/micro-banking/internal/app/auth"
	"github.com/tizzhh/micro-banking/internal/config"
	"github.com/tizzhh/micro-banking/internal/storage/postgres"
	"github.com/tizzhh/micro-banking/pkg/logger/sl"
)

func main() {
	cfg := config.Get()
	log := sl.Get()
	log.Info("starting auth app")

	storage, err := postgres.Get()

	if err != nil {
		panic(err)
	}

	authapp := authapp.New(log, cfg.GRPC.AuthPort, cfg.TokenTTL, storage)
	go authapp.GRPCServer.MustRun()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	<-stop

	authapp.GRPCServer.Stop()
	err = storage.Stop()
	if err != nil {
		log.Error("failed to stop storage", sl.Error(err))
	}
	log.Info("auth app stopped")
}
