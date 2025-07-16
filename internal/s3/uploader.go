package s3

import (
	"context"
	"fmt"
	"io"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type Uploader struct {
	client  *minio.Client
	bucket  string
	baseURL string
}

type Config struct {
	Endpoint  string
	AccessKey string
	SecretKey string
	Bucket    string
	UseSSL    bool
	BaseURL   string
}

type UploadInput struct {
	Filename    string
	Reader      io.Reader
	Size        int64
	ContentType string
}

func NewUploader(cfg Config) (*Uploader, error) {
	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		return nil, err
	}

	return &Uploader{
		client:  client,
		bucket:  cfg.Bucket,
		baseURL: cfg.BaseURL,
	}, nil
}

func (u *Uploader) Upload(ctx context.Context, input UploadInput) (string, error) {
	_, err := u.client.PutObject(ctx, u.bucket, input.Filename, input.Reader, input.Size, minio.PutObjectOptions{
		ContentType: input.ContentType,
	})
	if err != nil {
		return "", err
	}

	url := fmt.Sprintf("%s/%s", u.baseURL, input.Filename)
	return url, nil
}
