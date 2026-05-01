package file

import "io"

type FileInfo struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	NoteUUID    string `json:"note_uuid"`
	Size        int64  `json:"size"`
	ContentType string `json:"content_type"`
}

type DownloadFile struct {
	FileInfo
	Reader io.ReadCloser `json:"-"`
}

type UploadFile struct {
	ID          string
	Name        string
	NoteUUID    string
	Size        int64
	ContentType string
	Reader      io.Reader
}
