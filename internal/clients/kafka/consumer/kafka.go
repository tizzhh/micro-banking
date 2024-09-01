package consumer

import (
	"fmt"
	"log/slog"

	"github.com/IBM/sarama"
	"github.com/tizzhh/micro-banking/internal/config"
	"github.com/tizzhh/micro-banking/pkg/logger/sl"
)

type Consumer struct {
	log      *slog.Logger
	consumer sarama.Consumer
	topic    string
}

func New(log *slog.Logger, brokers []string, cfg config.KafkaConsumer, topic string) (*Consumer, error) {
	config := sarama.NewConfig()
	config.Consumer.Return.Errors = cfg.ReturnErrors

	consumer, err := sarama.NewConsumer(brokers, config)
	if err != nil {
		panic(err)
	}

	return &Consumer{
		log:      log,
		consumer: consumer,
		topic:    topic,
	}, nil
}

func (c *Consumer) Worker() (sarama.PartitionConsumer, error) {
	const caller = "clients.kafka.consumer.Worker"
	log := sl.AddCaller(c.log, caller)

	worker, err := c.consumer.ConsumePartition(c.topic, 0, sarama.OffsetOldest)
	if err != nil {
		log.Error("failed to get worker", sl.Error(err))
		return nil, fmt.Errorf("%s: %w", caller, err)
	}
	return worker, nil
}

func (c *Consumer) Stop() error {
	err := c.consumer.Close()
	if err != nil {
		return err
	}
	return nil
}
