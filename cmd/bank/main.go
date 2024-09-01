package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/tizzhh/micro-banking/internal/app/bank"
	"github.com/tizzhh/micro-banking/internal/config"
	"github.com/tizzhh/micro-banking/internal/storage/postgres"
	"github.com/tizzhh/micro-banking/pkg/logger/sl"
)

func main() {
	cfg := config.Get()
	log := sl.Get()

	log.Info("starting bank app")

	storage, err := postgres.Get()
	if err != nil {
		panic(err)
	}

	bankapp := bank.New(log, cfg, storage)

	go bankapp.HTTPServer.MustRun()
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	<-stop

	bankapp.HTTPServer.Stop()
	err = storage.Stop()
	if err != nil {
		log.Error("failed to stop storage", sl.Error(err))
	}
	log.Info("bank app stopped")
}
