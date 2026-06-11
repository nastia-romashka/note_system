package file

import "io"

type FileInfo struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	UserUUID    string `json:"user_uuid,omitempty"`
	WorkspaceID string `json:"workspace_id,omitempty"`
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
	UserUUID    string
	WorkspaceID string
	NoteUUID    string
	Size        int64
	ContentType string
	Reader      io.Reader
}

type FileStats struct {
	FilesCount int64 `json:"files_count"`
}
