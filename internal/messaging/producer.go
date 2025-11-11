package messaging

import (
	"context"
	"fmt"

	"telegram-normalizer/internal/logger"

	"github.com/IBM/sarama"
)

type Option struct {
	Logger       logger.Logger
	Broker       string
	SaslUsername string
	SaslPassword string
	ClientID     string
	Context      context.Context
}

type KafkaProducer struct {
	log      logger.Logger
	producer sarama.SyncProducer
}

func NewKafkaProducer(opt Option) (*KafkaProducer, error) {
	cfg := newProducerConfig(opt)
	ctx := ensureContext(opt.Context)

	var producer sarama.SyncProducer
	if err := connectWithRetry(ctx, opt.Logger, "Kafka producer", func() error {
		var err error
		producer, err = sarama.NewSyncProducer([]string{opt.Broker}, cfg)
		return err
	}); err != nil {
		return nil, err
	}

	if opt.Logger != nil {
		opt.Logger.Info("Kafka Producer ready: broker=%s", opt.Broker)
	}

	return &KafkaProducer{log: opt.Logger, producer: producer}, nil
}

func (kp *KafkaProducer) Send(_ context.Context, topic string, data []byte) error {
	if kp.producer == nil {
		return fmt.Errorf("produce: %w", ErrKafkaUnavailable)
	}

	msg := &sarama.ProducerMessage{Topic: topic, Value: sarama.ByteEncoder(data)}
	partition, offset, err := kp.producer.SendMessage(msg)
	if err != nil {
		if kp.log != nil {
			kp.log.Error("Kafka produce error: %v", err)
		}
		return fmt.Errorf("produce: %w", ErrKafkaUnavailable)
	}

	if kp.log != nil {
		kp.log.Info("Kafka delivered topic=%s partition=%d offset=%v bytes=%d", topic, partition, offset, len(data))
	}
	return nil
}

func (kp *KafkaProducer) Close() {
	if kp.producer == nil {
		return
	}
	_ = kp.producer.Close()
	if kp.log != nil {
		kp.log.Info("Kafka Producer closed")
	}
}
