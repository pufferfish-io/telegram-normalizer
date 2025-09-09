package tgnorm

import (
	"context"
	"encoding/json"
	"io"
	"mime"
	"path/filepath"

	"telegram-normalizer/internal/contract"
)

const defaultMimeType = "application/octet-stream"

type Producer interface {
	Send(_ context.Context, topic string, data []byte) error
}

type Parser interface {
	ParseTelegramMessage(data []byte) (contract.NormalizedMessage, error)
}

type Downloader interface {
	GetFilePath(fileID string) (string, error)
	DownloadFile(filePath string) (string, io.ReadSeeker, int64, error)
}

type Uploader interface {
	Upload(ctx context.Context, input contract.UploadFileRequest) (string, error)
}

type Option struct {
	KafkaTopic string
	Uploader   Uploader
	Parser     Parser
	Downloader Downloader
	Producer   Producer
}

type TelegramNormalizer struct {
	parser     Parser
	producer   Producer
	downloader Downloader
	uploader   Uploader
	kafkaTopic string
}

func NewTelegramNormalizer(opt Option) *TelegramNormalizer {
	return &TelegramNormalizer{
		parser:     opt.Parser,
		producer:   opt.Producer,
		downloader: opt.Downloader,
		uploader:   opt.Uploader,
		kafkaTopic: opt.KafkaTopic,
	}
}

func (t *TelegramNormalizer) Handle(ctx context.Context, raw []byte) error {
	msg, err := t.parser.ParseTelegramMessage(raw)
	if err != nil {
		return err
	}

	for i, media := range msg.Media {
		filePath, err := t.downloader.GetFilePath(media.OriginalFileID)
		if err != nil {
			return err
		}

		filename, reader, size, err := t.downloader.DownloadFile(filePath)
		if err != nil {
			return err
		}

		ext := filepath.Ext(filename)
		mimeType := mime.TypeByExtension(ext)
		if mimeType == "" {
			mimeType = defaultMimeType
		}

		s3URL, err := t.uploader.Upload(ctx, contract.UploadFileRequest{
			Filename:    filename,
			Reader:      reader,
			Size:        size,
			ContentType: mimeType,
		})
		if err != nil {
			return err
		}

		msg.Media[i].S3URL = s3URL
	}

	out, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	return t.producer.Send(ctx, t.kafkaTopic, out)
}
