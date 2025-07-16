package model

type KafkaConfig struct {
	BootstrapServersValue       string `yaml:"bootstrap_servers_value"`
	TelegramNormalizerTopicName string `yaml:"telegram_normalizer_topic_name"`
	TelegramUpdatesTopicName    string `yaml:"telegram_updates_topic_name"`
	TelegramUpdatesGroupId      string `yaml:"telegram_updates_group_id"`
}

func (KafkaConfig) SectionName() string {
	return "kafka"
}
