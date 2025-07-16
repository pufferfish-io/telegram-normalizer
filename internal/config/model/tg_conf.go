package model

type TelegramConfig struct {
	Token string `yaml:"token"`
}

func (TelegramConfig) SectionName() string {
	return "telegram"
}
