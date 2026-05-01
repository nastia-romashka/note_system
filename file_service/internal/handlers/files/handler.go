package files

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/julienschmidt/httprouter"

	"file_service/internal/apperror"
	domainfile "file_service/internal/file"
	"file_service/pkg/logging"
)

const (
	filesURL = "/api/files"
	fileURL  = "/api/files/:id"
)

type FileService interface {
	GetFile(noteUUID, fileID string) (domainfile.DownloadFile, error)
	GetFilesByNoteUUID(noteUUID string) ([]domainfile.FileInfo, error)
	Create(file domainfile.UploadFile) (domainfile.FileInfo, error)
	Delete(noteUUID, fileID string) error
}

type Handler struct {
	Logger             logging.Logger
	MaxUploadSizeBytes int64
	FileService        FileService
}

func (h *Handler) Register(router *httprouter.Router) {
	router.HandlerFunc(http.MethodGet, fileURL, apperror.Middleware(h.GetFile))
	router.HandlerFunc(http.MethodGet, filesURL, apperror.Middleware(h.GetFilesByNoteUUID))
	router.HandlerFunc(http.MethodPost, filesURL, apperror.Middleware(h.CreateFile))
	router.HandlerFunc(http.MethodDelete, fileURL, apperror.Middleware(h.DeleteFile))
}

func (h *Handler) GetFile(w http.ResponseWriter, r *http.Request) error {
	noteUUID := r.URL.Query().Get("note_uuid")
	if noteUUID == "" {
		return apperror.BadRequestError("note_uuid query parameter is required")
	}

	params := r.Context().Value(httprouter.ParamsKey).(httprouter.Params)
	fileID := params.ByName("id")
	if fileID == "" {
		return apperror.BadRequestError("file id is required")
	}

	file, err := h.FileService.GetFile(noteUUID, fileID)
	if err != nil {
		return err
	}
	defer file.Reader.Close()

	contentType := file.ContentType
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", file.Name))
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Length", fmt.Sprintf("%d", file.Size))
	w.WriteHeader(http.StatusOK)

	if _, err = io.Copy(w, file.Reader); err != nil {
		h.Logger.Error("failed to proxy file download", "note_uuid", noteUUID, "file_id", fileID, "error", err)
		return fmt.Errorf("proxy file download: %w", err)
	}

	return nil
}

func (h *Handler) GetFilesByNoteUUID(w http.ResponseWriter, r *http.Request) error {
	noteUUID := r.URL.Query().Get("note_uuid")
	if noteUUID == "" {
		return apperror.BadRequestError("note_uuid query parameter is required")
	}

	files, err := h.FileService.GetFilesByNoteUUID(noteUUID)
	if err != nil {
		return err
	}

	data, err := json.Marshal(files)
	if err != nil {
		h.Logger.Error("failed to marshal files response", "note_uuid", noteUUID, "error", err)
		return fmt.Errorf("marshal files response: %w", err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
	return nil
}

func (h *Handler) CreateFile(w http.ResponseWriter, r *http.Request) error {
	r.Body = http.MaxBytesReader(w, r.Body, h.MaxUploadSizeBytes)
	if err := r.ParseMultipartForm(h.MaxUploadSizeBytes); err != nil {
		h.Logger.Warn("failed to parse multipart form", "error", err)
		return apperror.BadRequestError("can't parse multipart form")
	}
	if r.MultipartForm != nil {
		defer r.MultipartForm.RemoveAll()
	}

	noteUUID := r.FormValue("note_uuid")
	if noteUUID == "" {
		return apperror.BadRequestError("note_uuid form field is required")
	}

	files, ok := r.MultipartForm.File["file"]
	if !ok || len(files) == 0 {
		return apperror.BadRequestError("file field is required")
	}

	fileHeader := files[0]
	reader, err := fileHeader.Open()
	if err != nil {
		h.Logger.Error("failed to open uploaded file", "file_name", fileHeader.Filename, "error", err)
		return fmt.Errorf("open uploaded file: %w", err)
	}
	defer reader.Close()

	created, err := h.FileService.Create(domainfile.UploadFile{
		Name:        fileHeader.Filename,
		NoteUUID:    noteUUID,
		Size:        fileHeader.Size,
		ContentType: fileHeader.Header.Get("Content-Type"),
		Reader:      reader,
	})
	if err != nil {
		return err
	}

	data, err := json.Marshal(created)
	if err != nil {
		h.Logger.Error("failed to marshal create file response", "note_uuid", noteUUID, "error", err)
		return fmt.Errorf("marshal create file response: %w", err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Location", fmt.Sprintf("%s/%s?note_uuid=%s", filesURL, created.ID, created.NoteUUID))
	w.WriteHeader(http.StatusCreated)
	_, _ = w.Write(data)
	return nil
}

func (h *Handler) DeleteFile(w http.ResponseWriter, r *http.Request) error {
	params := r.Context().Value(httprouter.ParamsKey).(httprouter.Params)
	fileID := params.ByName("id")
	if fileID == "" {
		return apperror.BadRequestError("file id is required")
	}

	noteUUID := r.URL.Query().Get("note_uuid")
	if noteUUID == "" {
		return apperror.BadRequestError("note_uuid query parameter is required")
	}

	if err := h.FileService.Delete(noteUUID, fileID); err != nil {
		return err
	}

	w.WriteHeader(http.StatusNoContent)
	return nil
}
