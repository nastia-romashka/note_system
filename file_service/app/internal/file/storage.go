package file

import "context"

type Storage interface {
	GetFile(ctx context.Context, noteUUID, fileID, userUUID string) (DownloadFile, error)
	GetFilesByNoteUUID(ctx context.Context, noteUUID, userUUID string) ([]FileInfo, error)
	CountFilesByUserUUID(ctx context.Context, userUUID string) (FileStats, error)
	CreateFile(ctx context.Context, file UploadFile) error
	DeleteFile(ctx context.Context, noteUUID, fileID, userUUID string) error
}
