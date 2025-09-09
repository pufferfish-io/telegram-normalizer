package telegram

import (
	"bytes"
	"fmt"
	"io"
	"path"

	"github.com/go-resty/resty/v2"
)

const NormalizedMessageDocType = "document"

const TelegramGetFileInfoUrl = "https://api.telegram.org/bot%s/getFile"
const TelegramDownloadFileUrl = "https://api.telegram.org/file/bot%s/%s"

type Downloader struct {
	Token string
	Api   *resty.Client
}

func NewDownloader(token string) *Downloader {
	return &Downloader{
		Token: token,
		Api:   resty.New(),
	}
}

func (d *Downloader) GetFilePath(fileID string) (string, error) {
	url := fmt.Sprintf(TelegramGetFileInfoUrl, d.Token)

	var result struct {
		OK     bool `json:"ok"`
		Result struct {
			FilePath string `json:"file_path"`
		} `json:"result"`
	}

	resp, err := d.Api.R().
		SetQueryParam("file_id", fileID).
		SetResult(&result).
		Get(url)

	if err != nil {
		return "", err
	}

	if !result.OK {
		return "", fmt.Errorf("telegram API error: %s", resp.String())
	}

	return result.Result.FilePath, nil
}

func (d *Downloader) DownloadFile(filePath string) (string, io.ReadSeeker, int64, error) {
	url := fmt.Sprintf(TelegramDownloadFileUrl, d.Token, filePath)

	resp, err := d.Api.R().Get(url)
	if err != nil {
		return "", nil, 0, err
	}

	if resp.StatusCode() != 200 {
		return "", nil, 0, fmt.Errorf("bad status: %d", resp.StatusCode())
	}

	filename := path.Base(filePath)
	data := resp.Body()
	reader := bytes.NewReader(data)

	return filename, reader, int64(len(data)), nil
}
