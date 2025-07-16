package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	cfg "telegram-normalizer/internal/config"
	cfgModel "telegram-normalizer/internal/config/model"
	"telegram-normalizer/internal/messaging"
	"telegram-normalizer/internal/processor"
	"telegram-normalizer/internal/s3"
)

func main() {
	log.Println("🚀 Starting telegram-normalizer...")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	kafkaConf, err := cfg.LoadSection[cfgModel.KafkaConfig]()
	if err != nil {
		log.Fatalf("❌ Failed to load Kafka config: %v", err)
	}

	s3Conf, err := cfg.LoadSection[cfgModel.S3Config]()
	if err != nil {
		log.Fatalf("S3 config error: %v", err)
	}

	tgConf, err := cfg.LoadSection[cfgModel.TelegramConfig]()
	if err != nil {
		log.Fatalf("TG config error: %v", err)
	}

	uploader, err := s3.NewUploader(s3.Config{
		Endpoint:  s3Conf.Endpoint,
		AccessKey: s3Conf.AccessKey,
		SecretKey: s3Conf.SecretKey,
		Bucket:    s3Conf.Bucket,
		UseSSL:    s3Conf.UseSSL,
		BaseURL:   s3Conf.BaseURL,
	})

	if err != nil {
		log.Fatalf("❌ Failed to create S3 uploader: %v", err)
	}

	normalizer := processor.NewTelegramNormalizer(tgConf.Token, kafkaConf.TelegramNormalizerTopicName, uploader)

	handler := func(msg []byte) error {
		log.Printf("🔧 Handle message: %s", string(msg))
		return normalizer.Handle(ctx, msg)
	}

	messaging.Init(kafkaConf.BootstrapServersValue)

	consumer, err := messaging.NewConsumer(kafkaConf.BootstrapServersValue, kafkaConf.TelegramUpdatesGroupId, kafkaConf.TelegramUpdatesTopicName, handler)
	if err != nil {
		log.Fatalf("❌ Failed to create consumer: %v", err)
	}

	go func() {
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
		<-ch
		cancel()
	}()

	if err := consumer.Start(ctx); err != nil {
		log.Fatalf("❌ Consumer error: %v", err)
	}
}
