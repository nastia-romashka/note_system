package file

import "context"

type Storage interface {
	GetFile(ctx context.Context, noteUUID, fileID, userUUID, workspaceID string) (DownloadFile, error)
	GetFilesByNoteUUID(ctx context.Context, noteUUID, userUUID, workspaceID string) ([]FileInfo, error)
	CountFiles(ctx context.Context, userUUID, workspaceID string) (FileStats, error)
	CreateFile(ctx context.Context, file UploadFile) error
	DeleteFile(ctx context.Context, noteUUID, fileID, userUUID, workspaceID string) error
	DeleteWorkspace(ctx context.Context, workspaceID string) error
}
