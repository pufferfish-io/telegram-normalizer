package model

type S3Config struct {
	Endpoint  string `yaml:"endpoint"`
	AccessKey string `yaml:"accesskey"`
	SecretKey string `yaml:"secretkey"`
	Bucket    string `yaml:"bucket"`
	UseSSL    bool   `yaml:"usessl"`
	BaseURL   string `yaml:"baseurl"`
}

func (S3Config) SectionName() string {
	return "s3"
}
