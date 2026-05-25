package file

import "io"

type UploadFileParams struct {
	NoteUUID    string
	UserUUID    string
	FileName    string
	ContentType string
	Size        int64
	Reader      io.Reader
}

type FileStats struct {
	FilesCount int64 `json:"files_count"`
}
