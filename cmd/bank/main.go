package main

import (
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/tizzhh/micro-banking/internal/app/bank"
	"github.com/tizzhh/micro-banking/internal/clients/kafka/producer"
	"github.com/tizzhh/micro-banking/internal/config"
	"github.com/tizzhh/micro-banking/internal/storage/postgres"
	"github.com/tizzhh/micro-banking/pkg/logger/sl"
)

const (
	KafkaTopic = "Mail"
)

// @title Micro-bank api
// @version 1.0
// @description This is a server for simulating some bank operations

// @contact.name tizzhh

// @host localhost:8080
// @BasePath /v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization

// @externalDocs.description  OpenAPI
// @externalDocs.url          https://swagger.io/resources/open-api/

func main() {
	cfg := config.Get()
	log := sl.Get()

	log.Info("starting bank app")

	storage, err := postgres.Get()
	if err != nil {
		panic(err)
	}

	brokers := strings.Split(cfg.Kafka.Brokers, ";")
	producer, err := producer.New(log, brokers, cfg.Kafka.Producer, KafkaTopic)
	if err != nil {
		panic(err)
	}

	bankapp := bank.New(log, cfg, storage, producer)

	go bankapp.HTTPServer.MustRun()
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	<-stop

	bankapp.HTTPServer.Stop()
	if err = storage.Stop(); err != nil {
		log.Error("failed to stop storage", sl.Error(err))
	}
	if err = producer.Stop(); err != nil {
		log.Error("failed to stop producer", sl.Error(err))
	}

	log.Info("bank app stopped")
}
