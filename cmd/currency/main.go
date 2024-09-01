package main

import (
	"os"
	"os/signal"
	"strings"
	"syscall"

	currencyapp "github.com/tizzhh/micro-banking/internal/app/currency"
	"github.com/tizzhh/micro-banking/internal/clients/kafka/producer"
	"github.com/tizzhh/micro-banking/internal/config"
	"github.com/tizzhh/micro-banking/internal/storage/postgres"
	"github.com/tizzhh/micro-banking/pkg/logger/sl"
)

const (
	KafkaTopic = "Mail"
)

func main() {
	cfg := config.Get()
	log := sl.Get()
	log.Info("starting auth app")

	storage, err := postgres.Get()
	if err != nil {
		panic(err)
	}

	brokers := strings.Split(cfg.Kafka.Brokers, ";")
	producer, err := producer.New(log, brokers, cfg.Kafka.Producer, KafkaTopic)
	if err != nil {
		panic(err)
	}

	currencyApp := currencyapp.New(log, cfg.GRPC.CurrencyPort, cfg.Redis.PingTimeout, cfg.CurrencyApi.Timeout, storage, producer)
	go currencyApp.GRPCServer.MustRun()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	<-stop

	currencyApp.GRPCServer.Stop()
	if err = storage.Stop(); err != nil {
		log.Error("failed to stop storage", sl.Error(err))
	}
	if err = producer.Stop(); err != nil {
		log.Error("failed to stop producer", sl.Error(err))
	}

	log.Info("auth app stopped")
}
