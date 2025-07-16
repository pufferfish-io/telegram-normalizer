package model

import (
	"time"
)

type NormalizedMessage struct {
	Source         string        `json:"source"`
	UserID         int64         `json:"user_id"`
	Username       *string       `json:"username,omitempty"`
	ChatID         int64         `json:"chat_id"`
	Text           *string       `json:"text,omitempty"`
	Timestamp      time.Time     `json:"timestamp"`
	Media          []MediaObject `json:"media,omitempty"`
	OriginalUpdate any           `json:"original_update"`
}

type MediaObject struct {
	Type           string  `json:"type"`
	OriginalFileID string  `json:"original_file_id"`
	Filename       *string `json:"filename,omitempty"`
	MimeType       *string `json:"mime_type,omitempty"`
	S3URL          string  `json:"s3_url"`
}
