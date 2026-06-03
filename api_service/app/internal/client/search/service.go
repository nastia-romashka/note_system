package search

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"myproject/internal/apperror"
	"myproject/pkg/logging"
	"myproject/pkg/rest"
)

type client struct {
	base rest.BaseClient
}

func NewService(baseURL string, logger logging.Logger) SearchService {
	return &client{
		base: rest.BaseClient{
			BaseURL: baseURL,
			HTTPClient: &http.Client{
				Timeout: 10 * time.Second,
			},
			Logger: logger,
		},
	}
}

type SearchService interface {
	SearchNotes(ctx context.Context, q, userUUID, categoryUUID string, tagUUIDs []string) ([]byte, error)
	UpsertNote(ctx context.Context, note IndexedNote) error
	UpsertNotes(ctx context.Context, notes []IndexedNote) error
	DeleteNote(ctx context.Context, noteUUID, userUUID string) error
	DeleteNotesByUser(ctx context.Context, userUUID string) error
}

func (c *client) SearchNotes(ctx context.Context, q, userUUID, categoryUUID string, tagUUIDs []string) ([]byte, error) {
	filters := []rest.FilterOptions{
		{Field: "user_uuid", Values: []string{userUUID}},
	}
	if q != "" {
		filters = append(filters, rest.FilterOptions{Field: "q", Values: []string{q}})
	}
	if categoryUUID != "" {
		filters = append(filters, rest.FilterOptions{Field: "category_uuid", Values: []string{categoryUUID}})
	}
	cleanTagUUIDs := make([]string, 0, len(tagUUIDs))
	for _, tagUUID := range tagUUIDs {
		if tagUUID != "" {
			cleanTagUUIDs = append(cleanTagUUIDs, tagUUID)
		}
	}
	if len(cleanTagUUIDs) > 0 {
		filters = append(filters, rest.FilterOptions{Field: "tag_uuid", Values: cleanTagUUIDs})
	}

	uri, err := c.base.BuildURL("search/notes", filters)
	if err != nil {
		return nil, fmt.Errorf("failed to build URL. error: %w", err)
	}

	req, err := http.NewRequest(http.MethodGet, uri, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create new request due to error: %w", err)
	}

	reqCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	req = req.WithContext(reqCtx)

	response, err := c.base.SendRequest(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request due to error: %w", err)
	}

	if response.IsOk {
		body, err := response.ReadBody()
		if err != nil {
			return nil, fmt.Errorf("failed to read body: %w", err)
		}
		return body, nil
	}

	return nil, apperror.APIError(
		response.StatusCode(),
		response.Error.ErrorCode,
		response.Error.Message,
		response.Error.DeveloperMessage,
	)
}

func (c *client) UpsertNote(ctx context.Context, note IndexedNote) error {
	return c.sendJSON(ctx, http.MethodPost, "index/notes", note, 5*time.Second)
}

func (c *client) UpsertNotes(ctx context.Context, notes []IndexedNote) error {
	if len(notes) == 0 {
		return nil
	}

	return c.sendJSON(ctx, http.MethodPost, "index/notes/import", notes, 20*time.Second)
}

func (c *client) DeleteNote(ctx context.Context, noteUUID, userUUID string) error {
	uri, err := c.base.BuildURL(fmt.Sprintf("index/notes/%s", noteUUID), []rest.FilterOptions{
		{Field: "user_uuid", Values: []string{userUUID}},
	})
	if err != nil {
		return fmt.Errorf("failed to build URL. error: %w", err)
	}

	req, err := http.NewRequest(http.MethodDelete, uri, nil)
	if err != nil {
		return fmt.Errorf("failed to create new request due to error: %w", err)
	}

	reqCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	req = req.WithContext(reqCtx)

	response, err := c.base.SendRequest(req)
	if err != nil {
		return fmt.Errorf("failed to send request due to error: %w", err)
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

func (c *client) DeleteNotesByUser(ctx context.Context, userUUID string) error {
	uri, err := c.base.BuildURL("index/notes", []rest.FilterOptions{
		{Field: "user_uuid", Values: []string{userUUID}},
	})
	if err != nil {
		return fmt.Errorf("failed to build URL. error: %w", err)
	}

	req, err := http.NewRequest(http.MethodDelete, uri, nil)
	if err != nil {
		return fmt.Errorf("failed to create new request due to error: %w", err)
	}

	reqCtx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()
	req = req.WithContext(reqCtx)

	response, err := c.base.SendRequest(req)
	if err != nil {
		return fmt.Errorf("failed to send request due to error: %w", err)
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

func (c *client) sendJSON(ctx context.Context, method, resource string, payload any, timeout time.Duration) error {
	uri, err := c.base.BuildURL(resource, nil)
	if err != nil {
		return fmt.Errorf("failed to build URL. error: %w", err)
	}

	dataBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequest(method, uri, bytes.NewBuffer(dataBytes))
	if err != nil {
		return fmt.Errorf("failed to create new request due to error: %w", err)
	}

	reqCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	req = req.WithContext(reqCtx)

	response, err := c.base.SendRequest(req)
	if err != nil {
		return fmt.Errorf("failed to send request due to error: %w", err)
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
