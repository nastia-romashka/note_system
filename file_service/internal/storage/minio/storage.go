package minio

import (
	"context"
	"errors"
	"fmt"

	miniosdk "github.com/minio/minio-go/v7"

	"file_service/internal/apperror"
	domainfile "file_service/internal/file"
	"file_service/pkg/logging"
	minioclient "file_service/pkg/minio"
)

type storage struct {
	client *minioclient.Client
	logger logging.Logger
}

func NewStorage(endpoint, accessKeyID, secretAccessKey string, useSSL bool, logger logging.Logger) (domainfile.Storage, error) {
	client, err := minioclient.NewClient(endpoint, accessKeyID, secretAccessKey, useSSL, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create minio client: %w", err)
	}

	return &storage{
		client: client,
		logger: logger,
	}, nil
}

func (s *storage) GetFile(ctx context.Context, noteUUID, fileID string) (result domainfile.DownloadFile, err error) {
	object, err := s.client.GetFile(ctx, noteUUID, fileID)
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
			ID:          info.Key,
			Name:        getObjectName(info),
			NoteUUID:    noteUUID,
			Size:        info.Size,
			ContentType: getContentType(info),
		},
		Reader: object,
	}, nil
}

func (s *storage) GetFilesByNoteUUID(ctx context.Context, noteUUID string) (result []domainfile.FileInfo, err error) {
	objects, err := s.client.ListFiles(ctx, noteUUID)
	if err != nil {
		if errors.Is(err, apperror.ErrNotFound) {
			return []domainfile.FileInfo{}, nil
		}
		response := miniosdk.ToErrorResponse(err)
		if response.Code == "NoSuchBucket" {
			return []domainfile.FileInfo{}, nil
		}
		return nil, wrapMinioError(err, "files not found")
	}

	files := make([]domainfile.FileInfo, 0, len(objects))
	for _, object := range objects {
		files = append(files, domainfile.FileInfo{
			ID:          object.Key,
			Name:        object.Name,
			NoteUUID:    noteUUID,
			Size:        object.Size,
			ContentType: object.ContentType,
		})
	}

	return files, nil
}

func (s *storage) CreateFile(ctx context.Context, file domainfile.UploadFile) (err error) {
	if err = s.client.EnsureBucket(ctx, file.NoteUUID); err != nil {
		return wrapMinioError(err, "failed to ensure note bucket")
	}
	if err = s.client.UploadFile(ctx, file.NoteUUID, file.ID, file.Name, file.ContentType, file.Size, file.Reader); err != nil {
		return wrapMinioError(err, "failed to upload file")
	}

	return nil
}

func (s *storage) DeleteFile(ctx context.Context, noteUUID, fileID string) (err error) {
	if _, err = s.client.StatFile(ctx, noteUUID, fileID); err != nil {
		return wrapMinioError(err, "file not found")
	}
	if err = s.client.DeleteFile(ctx, noteUUID, fileID); err != nil {
		return wrapMinioError(err, "failed to delete file")
	}

	return nil
}

func getObjectName(info miniosdk.ObjectInfo) string {
	if info.UserMetadata["Name"] != "" {
		return info.UserMetadata["Name"]
	}
	return info.Key
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
