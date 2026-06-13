package minio

import (
	"context"
	"errors"
	"fmt"
	"strings"

	miniosdk "github.com/minio/minio-go/v7"

	"file_service/internal/apperror"
	domainfile "file_service/internal/file"
	"file_service/pkg/logging"
	minioclient "file_service/pkg/minio"
)

type storage struct {
	client     *minioclient.Client
	logger     logging.Logger
	bucketName string
}

func NewStorage(endpoint, accessKeyID, secretAccessKey string, useSSL bool, bucketName string, logger logging.Logger) (domainfile.Storage, error) {
	client, err := minioclient.NewClient(endpoint, accessKeyID, secretAccessKey, useSSL, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create minio client: %w", err)
	}

	bucketName = strings.TrimSpace(bucketName)
	if bucketName == "" {
		return nil, fmt.Errorf("minio bucket is required")
	}

	return &storage{
		client:     client,
		logger:     logger,
		bucketName: bucketName,
	}, nil
}

func (s *storage) GetFile(ctx context.Context, noteUUID, fileID, _ string, workspaceID string) (domainfile.DownloadFile, error) {
	return s.getSharedFile(ctx, noteUUID, fileID, workspaceID)
}

func (s *storage) GetFilesByNoteUUID(ctx context.Context, noteUUID, _ string, workspaceID string) ([]domainfile.FileInfo, error) {
	return s.getSharedFilesByNote(ctx, noteUUID, workspaceID)
}

func (s *storage) CountFiles(ctx context.Context, _ string, workspaceID string) (stats domainfile.FileStats, err error) {
	objects, err := s.client.ListFiles(ctx, s.bucketName, workspacePrefix(workspaceID))
	if err != nil {
		return domainfile.FileStats{}, wrapMinioError(err, "files not found")
	}
	stats.FilesCount = int64(len(objects))
	return stats, nil
}

func (s *storage) CreateFile(ctx context.Context, file domainfile.UploadFile) error {
	if err := s.client.EnsureBucket(ctx, s.bucketName); err != nil {
		return wrapMinioError(err, "failed to ensure files bucket")
	}

	if err := s.client.UploadFile(
		ctx,
		s.bucketName,
		sharedObjectKey(file.WorkspaceID, file.NoteUUID, file.ID),
		file.Name,
		file.UserUUID,
		file.WorkspaceID,
		file.NoteUUID,
		file.ContentType,
		file.Size,
		file.Reader,
	); err != nil {
		return wrapMinioError(err, "failed to upload file")
	}

	return nil
}

func (s *storage) DeleteFile(ctx context.Context, noteUUID, fileID, _ string, workspaceID string) error {
	return s.deleteSharedFile(ctx, noteUUID, fileID, workspaceID)
}

func (s *storage) DeleteWorkspace(ctx context.Context, workspaceID string) error {
	objects, err := s.client.ListFiles(ctx, s.bucketName, workspacePrefix(workspaceID))
	if err != nil {
		response := miniosdk.ToErrorResponse(err)
		if response.Code == "NoSuchBucket" {
			return nil
		}
		return wrapMinioError(err, "workspace files not found")
	}

	for _, object := range objects {
		if err = s.client.DeleteFile(ctx, s.bucketName, object.Key); err != nil {
			return wrapMinioError(err, "failed to delete workspace files")
		}
	}

	return nil
}

func (s *storage) getSharedFile(ctx context.Context, noteUUID, fileID, workspaceID string) (domainfile.DownloadFile, error) {
	key := sharedObjectKey(workspaceID, noteUUID, fileID)
	object, err := s.client.GetFile(ctx, s.bucketName, key)
	if err != nil {
		return domainfile.DownloadFile{}, wrapMinioError(err, "file not found")
	}

	info, err := object.Stat()
	if err != nil {
		_ = object.Close()
		return domainfile.DownloadFile{}, wrapMinioError(err, "file not found")
	}

	return domainfile.DownloadFile{
		FileInfo: domainfile.FileInfo{
			ID:          fileID,
			Name:        getObjectName(info),
			UserUUID:    getObjectUserUUID(info),
			WorkspaceID: workspaceID,
			NoteUUID:    noteUUID,
			Size:        info.Size,
			ContentType: getContentType(info),
		},
		Reader: object,
	}, nil
}

func (s *storage) getSharedFilesByNote(ctx context.Context, noteUUID, workspaceID string) ([]domainfile.FileInfo, error) {
	objects, err := s.client.ListFiles(ctx, s.bucketName, notePrefix(workspaceID, noteUUID))
	if err != nil {
		response := miniosdk.ToErrorResponse(err)
		if response.Code == "NoSuchBucket" {
			return []domainfile.FileInfo{}, nil
		}
		return nil, wrapMinioError(err, "files not found")
	}

	files := make([]domainfile.FileInfo, 0, len(objects))
	for _, object := range objects {
		files = append(files, domainfile.FileInfo{
			ID:          sharedFileID(object.Key),
			Name:        object.Name,
			UserUUID:    object.UserUUID,
			WorkspaceID: workspaceID,
			NoteUUID:    noteUUID,
			Size:        object.Size,
			ContentType: object.ContentType,
		})
	}

	return files, nil
}

func (s *storage) deleteSharedFile(ctx context.Context, noteUUID, fileID, workspaceID string) error {
	key := sharedObjectKey(workspaceID, noteUUID, fileID)
	if _, err := s.client.StatFile(ctx, s.bucketName, key); err != nil {
		return wrapMinioError(err, "file not found")
	}
	if err := s.client.DeleteFile(ctx, s.bucketName, key); err != nil {
		return wrapMinioError(err, "failed to delete file")
	}

	return nil
}

func sharedObjectKey(workspaceID, noteUUID, fileID string) string {
	return fmt.Sprintf("workspace/%s/note/%s/%s", workspaceID, noteUUID, fileID)
}

func notePrefix(workspaceID, noteUUID string) string {
	return fmt.Sprintf("workspace/%s/note/%s/", workspaceID, noteUUID)
}

func workspacePrefix(workspaceID string) string {
	return fmt.Sprintf("workspace/%s/", workspaceID)
}

func sharedFileID(key string) string {
	parts := strings.Split(strings.TrimSpace(key), "/")
	if len(parts) == 0 {
		return key
	}
	return parts[len(parts)-1]
}

func getObjectName(info miniosdk.ObjectInfo) string {
	if info.UserMetadata["Name"] != "" {
		return info.UserMetadata["Name"]
	}
	return info.Key
}

func getObjectUserUUID(info miniosdk.ObjectInfo) string {
	return info.UserMetadata["User-Uuid"]
}

func getContentType(info miniosdk.ObjectInfo) string {
	if info.ContentType != "" {
		return info.ContentType
	}
	return "application/octet-stream"
}

func wrapMinioError(err error, notFoundMessage string) error {
	if err == nil {
		return nil
	}

	if errors.Is(err, apperror.ErrNotFound) {
		return apperror.NotFoundError(notFoundMessage)
	}

	response := miniosdk.ToErrorResponse(err)
	switch response.Code {
	case "NoSuchBucket", "NoSuchKey", "NoSuchObject":
		return apperror.NotFoundError(notFoundMessage)
	default:
		return err
	}
}
