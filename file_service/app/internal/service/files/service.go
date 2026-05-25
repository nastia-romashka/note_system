package files

import (
	"fmt"
	"mime"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/google/uuid"

	"file_service/internal/apperror"
	domainfile "file_service/internal/file"
	"file_service/pkg/logging"
)

var bucketNamePattern = regexp.MustCompile(`^[a-z0-9][a-z0-9.-]{1,61}[a-z0-9]$`)

type service struct {
	storage          domainfile.Storage
	logger           logging.Logger
	maxFileSizeBytes int64
}

func NewService(storage domainfile.Storage, logger logging.Logger, maxFileSizeBytes int64) *service {
	return &service{
		storage:          storage,
		logger:           logger,
		maxFileSizeBytes: maxFileSizeBytes,
	}
}

func (s *service) GetFile(noteUUID, fileID, userUUID string) (file domainfile.DownloadFile, err error) {
	noteUUID, err = validateNoteUUID(noteUUID)
	if err != nil {
		return domainfile.DownloadFile{}, err
	}
	userUUID, err = validateUserUUID(userUUID)
	if err != nil {
		return domainfile.DownloadFile{}, err
	}
	fileID = strings.TrimSpace(fileID)
	if fileID == "" {
		return domainfile.DownloadFile{}, apperror.BadRequestError("file id is required")
	}

	file, err = s.storage.GetFile(nilContext(), noteUUID, fileID, userUUID)
	if err != nil {
		return domainfile.DownloadFile{}, fmt.Errorf("get file: %w", err)
	}

	return file, nil
}

func (s *service) GetFilesByNoteUUID(noteUUID, userUUID string) (files []domainfile.FileInfo, err error) {
	noteUUID, err = validateNoteUUID(noteUUID)
	if err != nil {
		return nil, err
	}
	userUUID, err = validateUserUUID(userUUID)
	if err != nil {
		return nil, err
	}

	files, err = s.storage.GetFilesByNoteUUID(nilContext(), noteUUID, userUUID)
	if err != nil {
		return nil, fmt.Errorf("get files by note uuid: %w", err)
	}

	return files, nil
}

func (s *service) GetStats(userUUID string) (stats domainfile.FileStats, err error) {
	userUUID, err = validateUserUUID(userUUID)
	if err != nil {
		return domainfile.FileStats{}, err
	}

	stats, err = s.storage.CountFilesByUserUUID(nilContext(), userUUID)
	if err != nil {
		return domainfile.FileStats{}, fmt.Errorf("get file stats: %w", err)
	}

	return stats, nil
}

func (s *service) Create(file domainfile.UploadFile) (created domainfile.FileInfo, err error) {
	noteUUID, err := validateNoteUUID(file.NoteUUID)
	if err != nil {
		return domainfile.FileInfo{}, err
	}
	userUUID, err := validateUserUUID(file.UserUUID)
	if err != nil {
		return domainfile.FileInfo{}, err
	}
	if file.Reader == nil {
		return domainfile.FileInfo{}, apperror.BadRequestError("file reader is required")
	}
	if file.Size <= 0 {
		return domainfile.FileInfo{}, apperror.BadRequestError("file is empty")
	}
	if s.maxFileSizeBytes > 0 && file.Size > s.maxFileSizeBytes {
		return domainfile.FileInfo{}, apperror.BadRequestError("file is too large")
	}

	fileName := sanitizeFileName(file.Name)
	if fileName == "" {
		return domainfile.FileInfo{}, apperror.BadRequestError("file name is required")
	}

	fileID := uuid.NewString()
	contentType := sanitizeContentType(file.ContentType)

	file.ID = fileID
	file.Name = fileName
	file.NoteUUID = noteUUID
	file.UserUUID = userUUID
	file.ContentType = contentType

	if err = s.storage.CreateFile(nilContext(), file); err != nil {
		return domainfile.FileInfo{}, fmt.Errorf("create file: %w", err)
	}

	s.logger.Info("file uploaded", "note_uuid", noteUUID, "file_id", fileID, "file_name", fileName, "size", file.Size)
	return domainfile.FileInfo{
		ID:          fileID,
		Name:        fileName,
		UserUUID:    userUUID,
		NoteUUID:    noteUUID,
		Size:        file.Size,
		ContentType: contentType,
	}, nil
}

func (s *service) Delete(noteUUID, fileID, userUUID string) (err error) {
	noteUUID, err = validateNoteUUID(noteUUID)
	if err != nil {
		return err
	}
	userUUID, err = validateUserUUID(userUUID)
	if err != nil {
		return err
	}
	fileID = strings.TrimSpace(fileID)
	if fileID == "" {
		return apperror.BadRequestError("file id is required")
	}

	if err = s.storage.DeleteFile(nilContext(), noteUUID, fileID, userUUID); err != nil {
		return fmt.Errorf("delete file: %w", err)
	}

	s.logger.Info("file deleted", "note_uuid", noteUUID, "file_id", fileID)
	return nil
}

func sanitizeFileName(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return ""
	}

	name = filepath.Base(name)
	if name == "." || name == string(filepath.Separator) {
		return ""
	}

	return strings.ReplaceAll(name, "\x00", "")
}

func sanitizeContentType(contentType string) string {
	contentType = strings.TrimSpace(contentType)
	if contentType == "" {
		return "application/octet-stream"
	}

	mediaType, _, err := mime.ParseMediaType(contentType)
	if err != nil || mediaType == "" {
		return "application/octet-stream"
	}

	return mediaType
}

func validateNoteUUID(noteUUID string) (string, error) {
	noteUUID = strings.ToLower(strings.TrimSpace(noteUUID))
	if noteUUID == "" {
		return "", apperror.BadRequestError("note_uuid is required")
	}
	if !bucketNamePattern.MatchString(noteUUID) {
		return "", apperror.BadRequestError("note_uuid is not a valid bucket name")
	}

	return noteUUID, nil
}

func validateUserUUID(userUUID string) (string, error) {
	userUUID = strings.TrimSpace(userUUID)
	if userUUID == "" {
		return "", apperror.BadRequestError("user_uuid is required")
	}

	return userUUID, nil
}
