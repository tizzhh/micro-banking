package main

import (
	"encoding/json"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/tizzhh/micro-banking/internal/clients/kafka"
	"github.com/tizzhh/micro-banking/internal/clients/kafka/consumer"
	"github.com/tizzhh/micro-banking/internal/config"
	"github.com/tizzhh/micro-banking/pkg/logger/sl"
	"github.com/tizzhh/micro-banking/pkg/mail"
)

const (
	KafkaTopic = "Mail"
)

func main() {
	log := sl.Get()
	cfg := config.Get()

	brokers := strings.Split(cfg.Kafka.Brokers, ";")

	log.Info("starting mail app")

	mailApp := mail.New(log)

	consumer, err := consumer.New(log, brokers, cfg.Kafka.Consumer, KafkaTopic)
	if err != nil {
		panic(err)
	}
	worker, err := consumer.Worker()
	if err != nil {
		panic(err)
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	done := make(chan bool)

	go func() {
		for {
			select {
			case err := <-worker.Errors():
				log.Error("consumer-worker error", sl.Error(err))
			case msg := <-worker.Messages():
				log.Info("recieved msg")
				var emailMessage kafka.Message
				emailBytes := msg.Value
				err := json.Unmarshal(emailBytes, &emailMessage)
				if err != nil {
					log.Error("failed to unmarshal kafka message", sl.Error(err))
				}
				err = mailApp.SendMail(emailMessage.Message, []string{emailMessage.EmailAddr})
				log.Info("successfully sent mail")
				if err != nil {
					log.Error("failed to send mail", sl.Error(err))
				}
			case <-stop:
				done <- true
			}
		}
	}()

	<-done

	err = worker.Close()
	if err != nil {
		log.Error("failed to close consumer-worker", sl.Error(err))
	}
	err = consumer.Stop()
	if err != nil {
		log.Error("failed to close consumer", sl.Error(err))
	}
	log.Info("stopping mail app")
}
