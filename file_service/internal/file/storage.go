package file

import "context"

type Storage interface {
	GetFile(ctx context.Context, noteUUID, fileID string) (DownloadFile, error)
	GetFilesByNoteUUID(ctx context.Context, noteUUID string) ([]FileInfo, error)
	CreateFile(ctx context.Context, file UploadFile) error
	DeleteFile(ctx context.Context, noteUUID, fileID string) error
}
