package s3

import (
	"context"
	"fmt"
	"path/filepath"

	"telegram-normalizer/internal/contract"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type Uploader struct {
	client *minio.Client
	bucket string
}

type Option struct {
	Endpoint  string
	AccessKey string
	SecretKey string
	Bucket    string
	UseSSL    bool
}

func NewUploader(opt Option) (*Uploader, error) {
	client, err := minio.New(opt.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(opt.AccessKey, opt.SecretKey, ""),
		Secure: opt.UseSSL,
	})
	if err != nil {
		return nil, err
	}

	return &Uploader{
		client: client,
		bucket: opt.Bucket,
	}, nil
}

func (u *Uploader) Upload(ctx context.Context, input contract.UploadFileRequest) (string, error) {
	ext := filepath.Ext(input.Filename)
	objectKey := fmt.Sprintf("%s%s", uuid.NewString(), ext)

	_, err := u.client.PutObject(ctx, u.bucket, objectKey, input.Reader, input.Size, minio.PutObjectOptions{
		ContentType: input.ContentType,
		UserMetadata: map[string]string{
			"original-filename": input.Filename,
		},
	})
	if err != nil {
		return "", err
	}

	return objectKey, nil
}
