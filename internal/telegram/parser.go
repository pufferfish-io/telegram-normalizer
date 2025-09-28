package telegram

import (
	"encoding/json"
	"fmt"
	"time"

	"telegram-normalizer/internal/contract"
)

const NormalizedMessageSource = "tg"
const NormalizedMessageImageType = "photo"

type Update struct {
	UpdateID int64   `json:"update_id"`
	Message  Message `json:"message"`
}

type Message struct {
	MessageID int64   `json:"message_id"`
	From      User    `json:"from"`
	Chat      Chat    `json:"chat"`
	Date      int64   `json:"date"`
	Text      *string `json:"text,omitempty"`

	Photo    []PhotoSize `json:"photo,omitempty"`
	Document *Document   `json:"document,omitempty"`
}

type User struct {
	ID       int64   `json:"id"`
	Username *string `json:"username,omitempty"`
}

type Chat struct {
	ID int64 `json:"id"`
}

type PhotoSize struct {
	FileID string `json:"file_id"`
}

type Document struct {
	FileID   string  `json:"file_id"`
	FileName *string `json:"file_name,omitempty"`
	MimeType *string `json:"mime_type,omitempty"`
}

type TgMessageParser struct {
}

func NewTgMessageParser() *TgMessageParser {
	return &TgMessageParser{}
}

func (*TgMessageParser) ParseTelegramMessage(data []byte) (contract.NormalizedMessage, error) {
	var update Update
	if err := json.Unmarshal(data, &update); err != nil {
		return contract.NormalizedMessage{}, fmt.Errorf("parsing telegram message: %w", err)
	}

	msg := update.Message

	n := contract.NormalizedMessage{
		Source:         NormalizedMessageSource,
		UserID:         msg.From.ID,
		Username:       msg.From.Username,
		ChatID:         msg.Chat.ID,
		Text:           msg.Text,
		Timestamp:      time.Unix(msg.Date, 0).UTC(),
		Media:          []contract.MediaObject{},
		OriginalUpdate: json.RawMessage(data),
	}

	if len(msg.Photo) > 0 {
		last := msg.Photo[len(msg.Photo)-1]
		n.Media = append(n.Media, contract.MediaObject{
			Type:           NormalizedMessageImageType,
			OriginalFileID: last.FileID,
		})
	}

	if msg.Document != nil {
		n.Media = append(n.Media, contract.MediaObject{
			Type:           NormalizedMessageDocType,
			OriginalFileID: msg.Document.FileID,
			Filename:       msg.Document.FileName,
			MimeType:       msg.Document.MimeType,
		})
	}

	return n, nil
}
