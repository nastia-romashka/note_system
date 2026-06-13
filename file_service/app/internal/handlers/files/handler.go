package files

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"file_service/internal/apperror"
	domainfile "file_service/internal/file"
	"file_service/pkg/logging"
)

const (
	filesURL = "/api/files"
	fileURL  = "/api/files/{id}"
	statsURL = "/api/stats"
	workspaceURL = "/api/workspaces/{workspace_id}"
)

type FileService interface {
	GetFile(noteUUID, fileID, userUUID, workspaceID string) (domainfile.DownloadFile, error)
	GetFilesByNoteUUID(noteUUID, userUUID, workspaceID string) ([]domainfile.FileInfo, error)
	GetStats(userUUID, workspaceID string) (domainfile.FileStats, error)
	Create(file domainfile.UploadFile) (domainfile.FileInfo, error)
	Delete(noteUUID, fileID, userUUID, workspaceID string) error
	DeleteWorkspace(workspaceID string) error
}

type Handler struct {
	Logger             logging.Logger
	MaxUploadSizeBytes int64
	FileService        FileService
}

func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET "+fileURL, apperror.Middleware(h.GetFile))
	mux.HandleFunc("GET "+filesURL, apperror.Middleware(h.GetFilesByNoteUUID))
	mux.HandleFunc("POST "+filesURL, apperror.Middleware(h.CreateFile))
	mux.HandleFunc("DELETE "+fileURL, apperror.Middleware(h.DeleteFile))
	mux.HandleFunc("DELETE "+workspaceURL, apperror.Middleware(h.DeleteWorkspace))
	mux.HandleFunc("GET "+statsURL, apperror.Middleware(h.GetStats))
}

func (h *Handler) GetFile(w http.ResponseWriter, r *http.Request) error {
	noteUUID := r.URL.Query().Get("note_uuid")
	if noteUUID == "" {
		return apperror.BadRequestError("note_uuid query parameter is required")
	}
	userUUID, err := userUUIDFromQuery(r)
	if err != nil {
		return err
	}
	workspaceID := workspaceIDFromQuery(r)

	fileID := r.PathValue("id")
	if fileID == "" {
		return apperror.BadRequestError("file id is required")
	}

	file, err := h.FileService.GetFile(noteUUID, fileID, userUUID, workspaceID)
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
	userUUID, err := userUUIDFromQuery(r)
	if err != nil {
		return err
	}
	workspaceID := workspaceIDFromQuery(r)

	files, err := h.FileService.GetFilesByNoteUUID(noteUUID, userUUID, workspaceID)
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

func (h *Handler) GetStats(w http.ResponseWriter, r *http.Request) error {
	userUUID, err := userUUIDFromQuery(r)
	if err != nil {
		return err
	}
	workspaceID := workspaceIDFromQuery(r)

	stats, err := h.FileService.GetStats(userUUID, workspaceID)
	if err != nil {
		return err
	}

	data, err := json.Marshal(stats)
	if err != nil {
		h.Logger.Error("failed to marshal file stats response", "user_uuid", userUUID, "error", err)
		return fmt.Errorf("marshal file stats response: %w", err)
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
	userUUID := r.FormValue("user_uuid")
	if userUUID == "" {
		return apperror.BadRequestError("user_uuid form field is required")
	}
	workspaceID := r.FormValue("workspace_id")
	if workspaceID == "" {
		return apperror.BadRequestError("workspace_id form field is required")
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
		UserUUID:    userUUID,
		WorkspaceID: workspaceID,
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
	w.Header().Set("Location", fmt.Sprintf("%s/%s?note_uuid=%s&workspace_id=%s", filesURL, created.ID, created.NoteUUID, created.WorkspaceID))
	w.WriteHeader(http.StatusCreated)
	_, _ = w.Write(data)
	return nil
}

func (h *Handler) DeleteFile(w http.ResponseWriter, r *http.Request) error {
	fileID := r.PathValue("id")
	if fileID == "" {
		return apperror.BadRequestError("file id is required")
	}

	noteUUID := r.URL.Query().Get("note_uuid")
	if noteUUID == "" {
		return apperror.BadRequestError("note_uuid query parameter is required")
	}
	userUUID, err := userUUIDFromQuery(r)
	if err != nil {
		return err
	}
	workspaceID := workspaceIDFromQuery(r)

	if err := h.FileService.Delete(noteUUID, fileID, userUUID, workspaceID); err != nil {
		return err
	}

	w.WriteHeader(http.StatusNoContent)
	return nil
}

func (h *Handler) DeleteWorkspace(w http.ResponseWriter, r *http.Request) error {
	workspaceID := r.PathValue("workspace_id")
	if workspaceID == "" {
		return apperror.BadRequestError("workspace_id is required")
	}

	if err := h.FileService.DeleteWorkspace(workspaceID); err != nil {
		return err
	}

	w.WriteHeader(http.StatusNoContent)
	return nil
}

func userUUIDFromQuery(r *http.Request) (string, error) {
	userUUID := r.URL.Query().Get("user_uuid")
	if userUUID == "" {
		return "", apperror.BadRequestError("user_uuid query parameter is required")
	}

	return userUUID, nil
}

func workspaceIDFromQuery(r *http.Request) string {
	return r.URL.Query().Get("workspace_id")
}
