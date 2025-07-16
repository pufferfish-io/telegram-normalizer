package messaging

import (
	"context"
	"log"

	kafka "github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

type MessageHandler func([]byte) error

type Consumer struct {
	consumer *kafka.Consumer
	handler  MessageHandler
	topic    string
}

func NewConsumer(brokers string, groupID string, topic string, handler MessageHandler) (*Consumer, error) {
	c, err := kafka.NewConsumer(&kafka.ConfigMap{
		bootstrapServersKey: brokers,
		groupIDKey:          groupID,
		enableAutoCommitKey: enableAutoCommitFalse,
		autoOffsetResetKey:  autoOffsetResetEarliest,
	})
	if err != nil {
		return nil, err
	}

	return &Consumer{
		consumer: c,
		handler:  handler,
		topic:    topic,
	}, nil
}

func (c *Consumer) Start(ctx context.Context) error {
	log.Printf("Subscribing to Kafka topic: %s", c.topic)

	if err := c.consumer.SubscribeTopics([]string{c.topic}, nil); err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			log.Println("👋 Consumer shutdown requested")
			return c.consumer.Close()

		default:
			ev := c.consumer.Poll(100)
			if ev == nil {
				continue
			}

			switch msg := ev.(type) {
			case *kafka.Message:
				log.Printf("📦 Received message [%s]: %s", msg.TopicPartition, string(msg.Value))

				if err := c.handler(msg.Value); err != nil {
					log.Printf("Handler error: %v", err)
					continue
				}

				_, err := c.consumer.CommitMessage(msg)
				if err != nil {
					log.Printf("Commit error: %v", err)
				}

			case kafka.Error:
				log.Printf("Kafka error: %v", msg)
			}
		}
	}
}
