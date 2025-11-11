package main

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"syscall"
	"time"

	"telegram-normalizer/internal/config"
	"telegram-normalizer/internal/logger"
	"telegram-normalizer/internal/messaging"
	"telegram-normalizer/internal/s3"
	"telegram-normalizer/internal/telegram"
	"telegram-normalizer/internal/tgnorm"
)

func main() {
	// init logger first to use it everywhere
	lg, cleanup := logger.NewZapLogger()
	defer cleanup()
	lg.Info("🚀 Starting telegram-normalizer…")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg, err := config.Load()
	if err != nil {
		lg.Error("❌ Failed to load config: %v", err)
		os.Exit(1)
	}

	uploader, err := s3.NewUploader(s3.Option{
		Endpoint:  cfg.S3.Endpoint,
		AccessKey: cfg.S3.AccessKey,
		SecretKey: cfg.S3.SecretKey,
		Bucket:    cfg.S3.Bucket,
		UseSSL:    cfg.S3.UseSSL,
	})
	if err != nil {
		lg.Error("❌ Failed to create S3 uploader: %v", err)
		os.Exit(1)
	}

	downloader := telegram.NewDownloader(cfg.Telegram.Token)

	producer, err := messaging.NewKafkaProducer(messaging.Option{
		Logger:       lg,
		Broker:       cfg.Kafka.BootstrapServersValue,
		SaslUsername: cfg.Kafka.SaslUsername,
		SaslPassword: cfg.Kafka.SaslPassword,
		ClientID:     cfg.Kafka.ClientID,
	})
	if err != nil {
		lg.Error("❌ Failed to create Kafka producer: %v", err)
		os.Exit(1)
	}

	parser := telegram.NewTgMessageParser()

	normalizer := tgnorm.NewTelegramNormalizer(tgnorm.Option{
		KafkaTopic: cfg.Kafka.NormalizerTopicName,
		Uploader:   uploader,
		Parser:     parser,
		Downloader: downloader,
		Producer:   producer,
	})

	consumer, err := messaging.NewKafkaConsumer(messaging.ConsumerOption{
		Logger:       lg,
		Broker:       cfg.Kafka.BootstrapServersValue,
		GroupID:      cfg.Kafka.GroupID,
		Topics:       []string{cfg.Kafka.TgMessTopicName},
		Handler:      normalizer,
		SaslUsername: cfg.Kafka.SaslUsername,
		SaslPassword: cfg.Kafka.SaslPassword,
		ClientID:     cfg.Kafka.ClientID,
	})
	if err != nil {
		lg.Error("❌ Failed to create consumer: %v", err)
		os.Exit(1)
	}

	go func() {
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
		<-ch
		cancel()
	}()

	runConsumerSupervisor(ctx, consumer, lg)
}

func runConsumerSupervisor(ctx context.Context, consumer *messaging.KafkaConsumer, lg logger.Logger) {
	wait := 1 * time.Second
	maxWait := 30 * time.Second
	for {
		err := consumer.Start(ctx)
		if err == nil {
			lg.Info("consumer finished without error")
			return
		}
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			lg.Info("consumer stopped: %v", err)
			return
		}
		lg.Error("consumer error: %v — retry in %s", err, wait)
		select {
		case <-time.After(wait):
			wait *= 2
			if wait > maxWait {
				wait = maxWait
			}
		case <-ctx.Done():
			lg.Info("context canceled while waiting to restart consumer: %v", ctx.Err())
			return
		}
	}
}
