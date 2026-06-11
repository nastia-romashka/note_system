package files

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"

	"myproject/internal/apperror"
	categoryclient "myproject/internal/client/category"
	fileclient "myproject/internal/client/file"
	noteclient "myproject/internal/client/note"
	searchclient "myproject/internal/client/search"
	userclient "myproject/internal/client/user"
	"myproject/internal/handlers/actionlog"
	"myproject/internal/requestctx"
	"myproject/internal/searchsync"
	"myproject/pkg/logging"
	"myproject/pkg/middleware/jwt"
	workspacemw "myproject/pkg/middleware/workspace"
)

const (
	noteFilesURL = "/api/notes/{uuid}/files"
	noteFileURL  = "/api/notes/{uuid}/files/{fileId}"
)

type Handler struct {
	Logger           logging.Logger
	CategoryService  categoryclient.CategoryService
	NoteService      noteclient.NoteService
	FileService      fileclient.FileService
	SearchService    searchclient.SearchService
	WorkspaceService userclient.UserService
	ActionRecorder   actionlog.Recorder
}

func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc(http.MethodGet+" "+noteFilesURL, jwt.JWTMiddleware(workspacemw.Middleware(h.WorkspaceService, apperror.Middleware(h.GetFilesByNote))))
	mux.HandleFunc(http.MethodPost+" "+noteFilesURL, jwt.JWTMiddleware(workspacemw.Middleware(h.WorkspaceService, apperror.Middleware(h.UploadFileToNote))))
	mux.HandleFunc(http.MethodGet+" "+noteFileURL, jwt.JWTMiddleware(workspacemw.Middleware(h.WorkspaceService, apperror.Middleware(h.DownloadNoteFile))))
	mux.HandleFunc(http.MethodDelete+" "+noteFileURL, jwt.JWTMiddleware(workspacemw.Middleware(h.WorkspaceService, apperror.Middleware(h.DeleteNoteFile))))
}

func (h *Handler) GetFilesByNote(w http.ResponseWriter, r *http.Request) error {
	noteUUID, err := getNoteUUID(r)
	if err != nil {
		return err
	}
	userUUID, err := h.userUUIDFromContext(r)
	if err != nil {
		return err
	}
	workspaceID, err := h.workspaceIDFromContext(r)
	if err != nil {
		return err
	}

	if _, err = h.NoteService.GetNote(r.Context(), noteUUID, userUUID, workspaceID); err != nil {
		return err
	}

	data, err := h.FileService.GetNoteFiles(r.Context(), noteUUID, userUUID, workspaceID)
	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
	return nil
}

func (h *Handler) UploadFileToNote(w http.ResponseWriter, r *http.Request) error {
	noteUUID, err := getNoteUUID(r)
	if err != nil {
		return err
	}
	userUUID, err := h.userUUIDFromContext(r)
	if err != nil {
		return err
	}
	workspaceID, err := h.workspaceIDFromContext(r)
	if err != nil {
		return err
	}

	if _, err = h.NoteService.GetNote(r.Context(), noteUUID, userUUID, workspaceID); err != nil {
		return err
	}

	if err = r.ParseMultipartForm(32 << 20); err != nil {
		return apperror.BadRequestError("can't parse multipart form")
	}
	if r.MultipartForm != nil {
		defer r.MultipartForm.RemoveAll()
	}

	files, ok := r.MultipartForm.File["file"]
	if !ok || len(files) == 0 {
		return apperror.BadRequestError("file field is required")
	}

	fileHeader := files[0]
	reader, err := fileHeader.Open()
	if err != nil {
		return fmt.Errorf("open uploaded file: %w", err)
	}
	defer reader.Close()

	data, location, err := h.FileService.UploadNoteFile(r.Context(), fileclient.UploadFileParams{
		NoteUUID:    noteUUID,
		UserUUID:    userUUID,
		WorkspaceID: workspaceID,
		FileName:    fileHeader.Filename,
		ContentType: fileHeader.Header.Get("Content-Type"),
		Size:        fileHeader.Size,
		Reader:      reader,
	})
	if err != nil {
		return err
	}

	fileID := fileIDFromLocation(location)
	if location != "" {
		if resolvedLocation, resolveErr := buildGatewayFileLocation(location, noteUUID); resolveErr == nil {
			w.Header().Set("Location", resolvedLocation)
		}
	}
	h.ActionRecorder.Record(r, userUUID, "file.uploaded", "file", fileID, map[string]any{
		"note_uuid":    noteUUID,
		"file_name":    fileHeader.Filename,
		"content_type": fileHeader.Header.Get("Content-Type"),
		"size":         fileHeader.Size,
	})
	h.syncIndexedNote(r, userUUID, noteUUID, "file_upload")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_, _ = w.Write(data)
	return nil
}

func (h *Handler) DownloadNoteFile(w http.ResponseWriter, r *http.Request) error {
	noteUUID, err := getNoteUUID(r)
	if err != nil {
		return err
	}
	userUUID, err := h.userUUIDFromContext(r)
	if err != nil {
		return err
	}
	workspaceID, err := h.workspaceIDFromContext(r)
	if err != nil {
		return err
	}

	if _, err = h.NoteService.GetNote(r.Context(), noteUUID, userUUID, workspaceID); err != nil {
		return err
	}

	fileID, err := getFileID(r)
	if err != nil {
		return err
	}

	response, err := h.FileService.DownloadNoteFile(r.Context(), noteUUID, fileID, userUUID, workspaceID)
	if err != nil {
		return err
	}
	defer response.Body().Close()

	if contentType := response.Header().Get("Content-Type"); contentType != "" {
		w.Header().Set("Content-Type", contentType)
	}
	if disposition := response.Header().Get("Content-Disposition"); disposition != "" {
		w.Header().Set("Content-Disposition", disposition)
	}
	if contentLength := response.Header().Get("Content-Length"); contentLength != "" {
		w.Header().Set("Content-Length", contentLength)
	}

	w.WriteHeader(http.StatusOK)
	if _, err = io.Copy(w, response.Body()); err != nil {
		return fmt.Errorf("proxy file download: %w", err)
	}

	return nil
}

func (h *Handler) DeleteNoteFile(w http.ResponseWriter, r *http.Request) error {
	noteUUID, err := getNoteUUID(r)
	if err != nil {
		return err
	}
	userUUID, err := h.userUUIDFromContext(r)
	if err != nil {
		return err
	}
	workspaceID, err := h.workspaceIDFromContext(r)
	if err != nil {
		return err
	}

	if _, err = h.NoteService.GetNote(r.Context(), noteUUID, userUUID, workspaceID); err != nil {
		return err
	}

	fileID, err := getFileID(r)
	if err != nil {
		return err
	}

	if err = h.FileService.DeleteNoteFile(r.Context(), noteUUID, fileID, userUUID, workspaceID); err != nil {
		return err
	}
	h.ActionRecorder.Record(r, userUUID, "file.deleted", "file", fileID, map[string]any{
		"note_uuid": noteUUID,
	})
	h.syncIndexedNote(r, userUUID, noteUUID, "file_delete")

	w.WriteHeader(http.StatusNoContent)
	return nil
}

func getNoteUUID(r *http.Request) (string, error) {
	noteUUID := strings.TrimSpace(r.PathValue("uuid"))
	if noteUUID == "" {
		return "", apperror.BadRequestError("empty note uuid")
	}

	return noteUUID, nil
}

func getFileID(r *http.Request) (string, error) {
	fileID := strings.TrimSpace(r.PathValue("fileId"))
	if fileID == "" {
		return "", apperror.BadRequestError("empty file id")
	}

	return fileID, nil
}

func buildGatewayFileLocation(rawLocation, noteUUID string) (string, error) {
	locationURL, err := url.Parse(rawLocation)
	if err != nil {
		return "", err
	}

	fileID := path.Base(locationURL.Path)
	if fileID == "" || fileID == "." || fileID == "/" {
		return "", fmt.Errorf("empty file id in location")
	}

	return fmt.Sprintf("/api/notes/%s/files/%s", noteUUID, fileID), nil
}

func fileIDFromLocation(rawLocation string) string {
	locationURL, err := url.Parse(rawLocation)
	if err != nil {
		return ""
	}

	fileID := path.Base(locationURL.Path)
	if fileID == "." || fileID == "/" {
		return ""
	}

	return fileID
}

func (h *Handler) userUUIDFromContext(r *http.Request) (string, error) {
	rawUserUUID := r.Context().Value("user_uuid")
	if rawUserUUID == nil {
		h.Logger.Error("there is no user_uuid in context")
		return "", apperror.MissingUserUUIDError()
	}

	userUUID, ok := rawUserUUID.(string)
	if !ok || userUUID == "" {
		h.Logger.Error("there is no user_uuid in context")
		return "", apperror.MissingUserUUIDError()
	}

	return userUUID, nil
}

func (h *Handler) workspaceIDFromContext(r *http.Request) (string, error) {
	return requestctx.WorkspaceID(r)
}

func (h *Handler) syncIndexedNote(r *http.Request, userUUID, noteUUID, action string) {
	workspaceID, err := h.workspaceIDFromContext(r)
	if err != nil {
		h.Logger.Warn("failed to resolve workspace for file search sync", "user_uuid", userUUID, "note_uuid", noteUUID, "action", action, "error", err)
		return
	}

	noteData, err := h.NoteService.GetNote(r.Context(), noteUUID, userUUID, workspaceID)
	if err != nil {
		h.Logger.Warn("failed to fetch note for file search sync", "user_uuid", userUUID, "note_uuid", noteUUID, "action", action, "error", err)
		return
	}

	var note noteclient.Note
	if err = json.Unmarshal(noteData, &note); err != nil {
		h.Logger.Warn("failed to decode note for file search sync", "user_uuid", userUUID, "note_uuid", noteUUID, "action", action, "error", err)
		return
	}

	categoryData, err := h.CategoryService.GetWorkspaceCategories(r.Context(), workspaceID)
	if err != nil {
		h.Logger.Warn("failed to fetch categories for file search sync", "user_uuid", userUUID, "note_uuid", noteUUID, "action", action, "error", err)
		return
	}

	var categories []categoryclient.Category
	if err = json.Unmarshal(categoryData, &categories); err != nil {
		h.Logger.Warn("failed to decode categories for file search sync", "user_uuid", userUUID, "note_uuid", noteUUID, "action", action, "error", err)
		return
	}

	tags, err := h.fetchTagsForNotes(r, userUUID, workspaceID, []noteclient.Note{note})
	if err != nil {
		h.Logger.Warn("failed to fetch tags for file search sync", "user_uuid", userUUID, "note_uuid", noteUUID, "action", action, "error", err)
		return
	}

	filesByNote, err := h.fetchFilesForNotes(r, userUUID, workspaceID, []noteclient.Note{note})
	if err != nil {
		h.Logger.Warn("failed to fetch files for file search sync", "user_uuid", userUUID, "note_uuid", noteUUID, "action", action, "error", err)
		return
	}

	document, err := searchsync.BuildIndexedNote(note, categories, tags, filesByNote[note.Uuid])
	if err != nil {
		h.Logger.Warn("failed to build indexed note for file search sync", "user_uuid", userUUID, "note_uuid", noteUUID, "action", action, "error", err)
		return
	}

	if err = h.SearchService.UpsertNote(r.Context(), document); err != nil {
		h.Logger.Warn("failed to sync note after file change", "user_uuid", userUUID, "note_uuid", noteUUID, "action", action, "error", err)
	}
}

func (h *Handler) fetchTagsForNotes(r *http.Request, userUUID, workspaceID string, notes []noteclient.Note) ([]noteclient.Tag, error) {
	tagUUIDs := searchsync.CollectTagUUIDs(notes)
	if len(tagUUIDs) == 0 {
		return nil, nil
	}

	tagsData, err := h.NoteService.GetTags(r.Context(), tagUUIDs, userUUID, workspaceID)
	if err != nil {
		return nil, err
	}

	var tags []noteclient.Tag
	if err = json.Unmarshal(tagsData, &tags); err != nil {
		return nil, fmt.Errorf("decode tags response: %w", err)
	}

	return tags, nil
}

func (h *Handler) fetchFilesForNotes(r *http.Request, userUUID, workspaceID string, notes []noteclient.Note) (map[string][]fileclient.FileInfo, error) {
	filesByNote := make(map[string][]fileclient.FileInfo, len(notes))

	for _, note := range notes {
		filesData, err := h.FileService.GetNoteFiles(r.Context(), note.Uuid, userUUID, workspaceID)
		if err != nil {
			return nil, fmt.Errorf("fetch files for note %s: %w", note.Uuid, err)
		}

		var files []fileclient.FileInfo
		if err = json.Unmarshal(filesData, &files); err != nil {
			return nil, fmt.Errorf("decode files response for note %s: %w", note.Uuid, err)
		}

		filesByNote[note.Uuid] = files
	}

	return filesByNote, nil
}
