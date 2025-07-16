package model

type ServerConfig struct {
	Port int `yaml:"port"`
}

func (ServerConfig) SectionName() string {
	return "server"
}
