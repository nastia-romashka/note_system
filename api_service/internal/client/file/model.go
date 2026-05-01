package file

import "io"

type UploadFileParams struct {
	NoteUUID    string
	FileName    string
	ContentType string
	Size        int64
	Reader      io.Reader
}
