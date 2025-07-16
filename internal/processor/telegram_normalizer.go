package processor

import (
	"context"
	"encoding/json"
	"mime"
	"path/filepath"

	"telegram-normalizer/internal/messaging"
	"telegram-normalizer/internal/model"
	"telegram-normalizer/internal/s3"
	"telegram-normalizer/internal/telegram"
)

const defaultMimeType = "application/octet-stream"

type TelegramNormalizer struct {
	parser     func([]byte) (*model.NormalizedMessage, error)
	downloader *telegram.Downloader
	uploader   *s3.Uploader
	kafkaTopic string
}

func NewTelegramNormalizer(tgToken string, kafkaTopic string, uploader *s3.Uploader) *TelegramNormalizer {
	return &TelegramNormalizer{
		parser:     telegram.ParseTelegramMessage,
		downloader: telegram.NewDownloader(tgToken),
		uploader:   uploader,
		kafkaTopic: kafkaTopic,
	}
}

func (t *TelegramNormalizer) Handle(ctx context.Context, raw []byte) error {
	msg, err := t.parser(raw)
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

		s3URL, err := t.uploader.Upload(ctx, s3.UploadInput{
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

	return messaging.Send(t.kafkaTopic, out)
}
