package messaging

import (
	"context"
	"errors"
	"fmt"
	"time"

	"telegram-normalizer/internal/logger"

	"github.com/IBM/sarama"
	"github.com/xdg-go/scram"
)

type Option struct {
	Logger       logger.Logger
	Broker       string
	SaslUsername string
	SaslPassword string
	ClientID     string
}

type KafkaProducer struct {
	log      logger.Logger
	producer sarama.SyncProducer
}

func NewKafkaProducer(opt Option) (*KafkaProducer, error) {
	cfg := sarama.NewConfig()

	cfg.ClientID = opt.ClientID
	cfg.Version = sarama.V2_8_0_0

	cfg.Producer.Return.Successes = true
	cfg.Producer.RequiredAcks = sarama.WaitForAll
	cfg.Producer.Retry.Max = 5
	cfg.Producer.Retry.Backoff = 100 * time.Millisecond

	cfg.Producer.Idempotent = true
	cfg.Net.MaxOpenRequests = 1

	cfg.Producer.Compression = sarama.CompressionSnappy
	cfg.Producer.Flush.Frequency = 50 * time.Millisecond
	cfg.Producer.Flush.Messages = 100

	cfg.Net.SASL.Enable = true
	cfg.Net.SASL.User = opt.SaslUsername
	cfg.Net.SASL.Password = opt.SaslPassword
	cfg.Net.SASL.Mechanism = sarama.SASLTypeSCRAMSHA512
	cfg.Net.SASL.SCRAMClientGeneratorFunc = func() sarama.SCRAMClient {
		return &xdgSCRAMClient{hash: scram.SHA512}
	}

	p, err := sarama.NewSyncProducer([]string{opt.Broker}, cfg)
	if err != nil {
		return nil, fmt.Errorf("kafka producer init: %w", err)
	}
	opt.Logger.Info("Kafka Producer init: broker=%s", opt.Broker)
	return &KafkaProducer{log: opt.Logger, producer: p}, nil
}

func (kp *KafkaProducer) Send(_ context.Context, topic string, data []byte) error {
	msg := &sarama.ProducerMessage{Topic: topic, Value: sarama.ByteEncoder(data)}
	partition, offset, err := kp.producer.SendMessage(msg)
	if err != nil {
		return fmt.Errorf("produce: %w", err)
	}
	kp.log.Debug("Kafka delivered topic=%s partition=%d offset=%v bytes=%d", topic, partition, offset, len(data))
	return nil
}

func (kp *KafkaProducer) Close() {
	if err := kp.producer.Close(); err != nil {
		kp.log.Error("Kafka Producer close error: %v", err)
		return
	}
	kp.log.Info("Kafka Producer closed")
}

type xdgSCRAMClient struct {
	hash scram.HashGeneratorFcn
	*scram.Client
	*scram.ClientConversation
}

func (x *xdgSCRAMClient) Begin(userName, password, authzID string) error {
	c, err := x.hash.NewClient(userName, password, authzID)
	if err != nil {
		return err
	}
	x.Client = c
	x.ClientConversation = c.NewConversation()
	return nil
}

func (x *xdgSCRAMClient) Step(challenge string) (string, error) {
	if x.ClientConversation == nil {
		return "", errors.New("no scram conversation")
	}
	return x.ClientConversation.Step(challenge)
}

func (x *xdgSCRAMClient) Done() bool {
	if x.ClientConversation == nil {
		return false
	}
	return x.ClientConversation.Done()
}
