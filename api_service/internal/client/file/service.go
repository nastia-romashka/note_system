package file

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"
	"time"

	"myproject/internal/apperror"
	"myproject/pkg/logging"
	"myproject/pkg/rest"
)

type client struct {
	base     rest.BaseClient
	Resource string
}

func NewService(baseURL string, resource string, logger logging.Logger) FileService {
	return &client{
		Resource: resource,
		base: rest.BaseClient{
			BaseURL: baseURL,
			HTTPClient: &http.Client{
				Timeout: 60 * time.Second,
			},
			Logger: logger,
		},
	}
}

type FileService interface {
	GetNoteFiles(ctx context.Context, noteUUID string) ([]byte, error)
	UploadNoteFile(ctx context.Context, params UploadFileParams) ([]byte, string, error)
	DownloadNoteFile(ctx context.Context, noteUUID, fileID string) (*rest.APIResponse, error)
	DeleteNoteFile(ctx context.Context, noteUUID, fileID string) error
}

func (c *client) GetNoteFiles(ctx context.Context, noteUUID string) ([]byte, error) {
	var files []byte

	uri, err := c.base.BuildURL(c.Resource, []rest.FilterOptions{
		{
			Field:  "note_uuid",
			Values: []string{noteUUID},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to build URL: %w", err)
	}

	req, err := http.NewRequest(http.MethodGet, uri, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	reqCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	req = req.WithContext(reqCtx)

	response, err := c.base.SendRequest(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	if response.IsOk {
		files, err = response.ReadBody()
		if err != nil {
			return nil, fmt.Errorf("failed to read body: %w", err)
		}
		return files, nil
	}

	return nil, apperror.APIError(
		response.StatusCode(),
		response.Error.ErrorCode,
		response.Error.Message,
		response.Error.DeveloperMessage,
	)
}

func (c *client) UploadNoteFile(ctx context.Context, params UploadFileParams) ([]byte, string, error) {
	uri, err := c.base.BuildURL(c.Resource, nil)
	if err != nil {
		return nil, "", fmt.Errorf("failed to build URL: %w", err)
	}

	bodyReader, contentType, err := buildMultipartPayload(params)
	if err != nil {
		return nil, "", err
	}

	req, err := http.NewRequest(http.MethodPost, uri, bodyReader)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", contentType)

	reqCtx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()
	req = req.WithContext(reqCtx)

	response, err := c.base.SendRequest(req)
	if err != nil {
		return nil, "", fmt.Errorf("failed to send request: %w", err)
	}

	if response.IsOk {
		data, readErr := response.ReadBody()
		if readErr != nil {
			return nil, "", fmt.Errorf("failed to read body: %w", readErr)
		}

		location, locErr := response.Location()
		if locErr != nil {
			return data, "", nil
		}

		return data, location.String(), nil
	}

	return nil, "", apperror.APIError(
		response.StatusCode(),
		response.Error.ErrorCode,
		response.Error.Message,
		response.Error.DeveloperMessage,
	)
}

func (c *client) DownloadNoteFile(ctx context.Context, noteUUID, fileID string) (*rest.APIResponse, error) {
	uri, err := c.base.BuildURL(fmt.Sprintf("%s/%s", c.Resource, fileID), []rest.FilterOptions{
		{
			Field:  "note_uuid",
			Values: []string{noteUUID},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to build URL: %w", err)
	}

	req, err := http.NewRequest(http.MethodGet, uri, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Accept", "*/*")

	reqCtx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()
	req = req.WithContext(reqCtx)

	response, err := c.base.SendRequest(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	if response.IsOk {
		return response, nil
	}

	return nil, apperror.APIError(
		response.StatusCode(),
		response.Error.ErrorCode,
		response.Error.Message,
		response.Error.DeveloperMessage,
	)
}

func (c *client) DeleteNoteFile(ctx context.Context, noteUUID, fileID string) error {
	uri, err := c.base.BuildURL(fmt.Sprintf("%s/%s", c.Resource, fileID), []rest.FilterOptions{
		{
			Field:  "note_uuid",
			Values: []string{noteUUID},
		},
	})
	if err != nil {
		return fmt.Errorf("failed to build URL: %w", err)
	}

	req, err := http.NewRequest(http.MethodDelete, uri, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	reqCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	req = req.WithContext(reqCtx)

	response, err := c.base.SendRequest(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	if response.IsOk {
		return nil
	}

	return apperror.APIError(
		response.StatusCode(),
		response.Error.ErrorCode,
		response.Error.Message,
		response.Error.DeveloperMessage,
	)
}

func buildMultipartPayload(params UploadFileParams) (io.Reader, string, error) {
	if strings.TrimSpace(params.NoteUUID) == "" {
		return nil, "", apperror.BadRequestError("empty note uuid")
	}
	if params.Reader == nil {
		return nil, "", apperror.BadRequestError("file reader is required")
	}

	pipeReader, pipeWriter := io.Pipe()
	writer := multipart.NewWriter(pipeWriter)

	go func() {
		defer pipeWriter.Close()
		defer writer.Close()

		if err := writer.WriteField("note_uuid", params.NoteUUID); err != nil {
			_ = pipeWriter.CloseWithError(err)
			return
		}

		part, err := writer.CreateFormFile("file", params.FileName)
		if err != nil {
			_ = pipeWriter.CloseWithError(err)
			return
		}

		if _, err = io.Copy(part, params.Reader); err != nil {
			_ = pipeWriter.CloseWithError(err)
			return
		}
	}()

	return pipeReader, writer.FormDataContentType(), nil
}
