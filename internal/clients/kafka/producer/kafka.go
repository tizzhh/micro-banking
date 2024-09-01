package producer

import (
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/IBM/sarama"
	"github.com/tizzhh/micro-banking/internal/clients/kafka"
	"github.com/tizzhh/micro-banking/internal/config"
	"github.com/tizzhh/micro-banking/pkg/logger/sl"
)

type Producer struct {
	log      *slog.Logger
	producer sarama.SyncProducer
	topic    string
}

func New(log *slog.Logger, brokers []string, cfg config.KafkaProducer, topic string) (*Producer, error) {
	config := sarama.NewConfig()
	config.Producer.Return.Successes = cfg.ReturnSuccesses
	config.Producer.RequiredAcks = sarama.RequiredAcks(cfg.RequiredAcks)
	config.Producer.Retry.Max = cfg.RetryMax

	producer, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		panic(err)
	}

	return &Producer{
		log:      log,
		producer: producer,
		topic:    topic,
	}, nil
}

func (p *Producer) Produce(emailAddr string, msg string) error {
	const caller = "clients.kafka.producer.Produce"
	log := sl.AddCaller(p.log, caller)

	log.Info("producing message", slog.String("addr", emailAddr), slog.String("msg", msg))

	newMessage := &kafka.Message{
		EmailAddr: emailAddr,
		Message:   msg,
	}
	msgBytes, err := json.Marshal(newMessage)
	if err != nil {
		log.Error("failed to marshal msg", sl.Error(err))
		return fmt.Errorf("%s: %w", caller, err)
	}
	err = p.pushMessageToQueue(p.topic, msgBytes)
	if err != nil {
		log.Error("failed to push msg to queue", sl.Error(err))
		return fmt.Errorf("%s: %w", caller, err)
	}

	return nil
}

func (p *Producer) pushMessageToQueue(topic string, message []byte) error {
	const caller = "clients.kafka.producer.pushMessageToQueue"
	log := sl.AddCaller(p.log, caller)

	msg := &sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.StringEncoder(message),
	}

	log.Info("sending msg to queue")

	partition, offset, err := p.producer.SendMessage(msg)
	if err != nil {
		log.Error("failed to push msg to queue", sl.Error(err))
		return fmt.Errorf("%s: %w", caller, err)
	}
	log.Info("msg stored", slog.Int("partition", int(partition)), slog.Int("offset", int(offset)))

	return nil
}

func (p *Producer) Stop() error {
	err := p.producer.Close()
	if err != nil {
		return err
	}
	return nil
}
