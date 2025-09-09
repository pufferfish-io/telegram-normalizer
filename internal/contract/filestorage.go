package contract

import "io"

type UploadFileRequest struct {
	Filename    string
	Reader      io.Reader
	Size        int64
	ContentType string
}
