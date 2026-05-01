package files

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"

	"github.com/julienschmidt/httprouter"

	"myproject/internal/apperror"
	fileclient "myproject/internal/client/file"
	noteclient "myproject/internal/client/note"
	"myproject/pkg/logging"
	"myproject/pkg/middleware/jwt"
)

const (
	noteFilesURL = "/api/notes/:uuid/files"
	noteFileURL  = "/api/notes/:uuid/files/:fileId"
)

type Handler struct {
	Logger      logging.Logger
	NoteService noteclient.NoteService
	FileService fileclient.FileService
}

func (h *Handler) Register(router *httprouter.Router) {
	router.HandlerFunc(http.MethodGet, noteFilesURL, jwt.JWTMiddleware(apperror.Middleware(h.GetFilesByNote)))
	router.HandlerFunc(http.MethodPost, noteFilesURL, jwt.JWTMiddleware(apperror.Middleware(h.UploadFileToNote)))
	router.HandlerFunc(http.MethodGet, noteFileURL, jwt.JWTMiddleware(apperror.Middleware(h.DownloadNoteFile)))
	router.HandlerFunc(http.MethodDelete, noteFileURL, jwt.JWTMiddleware(apperror.Middleware(h.DeleteNoteFile)))
}

func (h *Handler) GetFilesByNote(w http.ResponseWriter, r *http.Request) error {
	noteUUID, err := getNoteUUID(r)
	if err != nil {
		return err
	}

	if _, err = h.NoteService.GetNote(r.Context(), noteUUID); err != nil {
		return err
	}

	data, err := h.FileService.GetNoteFiles(r.Context(), noteUUID)
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

	if _, err = h.NoteService.GetNote(r.Context(), noteUUID); err != nil {
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
		FileName:    fileHeader.Filename,
		ContentType: fileHeader.Header.Get("Content-Type"),
		Size:        fileHeader.Size,
		Reader:      reader,
	})
	if err != nil {
		return err
	}

	if location != "" {
		if resolvedLocation, resolveErr := buildGatewayFileLocation(location, noteUUID); resolveErr == nil {
			w.Header().Set("Location", resolvedLocation)
		}
	}
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

	if _, err = h.NoteService.GetNote(r.Context(), noteUUID); err != nil {
		return err
	}

	fileID, err := getFileID(r)
	if err != nil {
		return err
	}

	response, err := h.FileService.DownloadNoteFile(r.Context(), noteUUID, fileID)
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

	if _, err = h.NoteService.GetNote(r.Context(), noteUUID); err != nil {
		return err
	}

	fileID, err := getFileID(r)
	if err != nil {
		return err
	}

	if err = h.FileService.DeleteNoteFile(r.Context(), noteUUID, fileID); err != nil {
		return err
	}

	w.WriteHeader(http.StatusNoContent)
	return nil
}

func getNoteUUID(r *http.Request) (string, error) {
	params := r.Context().Value(httprouter.ParamsKey).(httprouter.Params)
	noteUUID := strings.TrimSpace(params.ByName("uuid"))
	if noteUUID == "" {
		return "", apperror.BadRequestError("empty note uuid")
	}

	return noteUUID, nil
}

func getFileID(r *http.Request) (string, error) {
	params := r.Context().Value(httprouter.ParamsKey).(httprouter.Params)
	fileID := strings.TrimSpace(params.ByName("fileId"))
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
