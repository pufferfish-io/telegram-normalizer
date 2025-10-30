package config

import (
	"fmt"

	"github.com/caarlos0/env/v11"
	"github.com/go-playground/validator/v10"
)

type Kafka struct {
	BootstrapServersValue string `validate:"required" env:"BOOTSTRAP_SERVERS_VALUE"`
	GroupID               string `validate:"required" env:"GROUP_ID_TELEGRAM_NORMALIZER"`
	TgMessTopicName       string `validate:"required" env:"TOPIC_NAME_TELEGRAM_UPDATES"`
	NormalizerTopicName   string `env:"TOPIC_NAME_NORMALIZED_MSG"`
	SaslUsername          string `env:"SASL_USERNAME"`
	SaslPassword          string `env:"SASL_PASSWORD"`
	ClientID              string `env:"CLIENT_ID_TELEGRAM_NORMALIZER"`
}

type Telegram struct {
	Token string `validate:"required" env:"TOKEN"`
}

type S3 struct {
	Endpoint  string `validate:"required" env:"ENDPOINT"`
	AccessKey string `validate:"required" env:"ACCESS_KEY"`
	SecretKey string `validate:"required" env:"SECRET_KEY"`
	Bucket    string `validate:"required" env:"BUCKET"`
	UseSSL    bool   `env:"USE_SSL"`
}

type Config struct {
	Telegram Telegram `envPrefix:"TELEGRAM"`
	Kafka    Kafka    `envPrefix:"KAFKA_"`
	S3       S3       `envPrefix:"S3_"`
}

func Load() (*Config, error) {
	var c Config
	if err := env.Parse(&c); err != nil {
		return nil, fmt.Errorf("env parse: %w", err)
	}
	v := validator.New()
	if err := v.Struct(c); err != nil {
		return nil, fmt.Errorf("config validate: %w", err)
	}
	return &c, nil
}
