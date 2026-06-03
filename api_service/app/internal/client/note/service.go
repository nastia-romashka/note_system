package note

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
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

func NewService(baseURL string, resource string, logger logging.Logger) NoteService {
	return &client{
		Resource: resource,
		base: rest.BaseClient{
			BaseURL: baseURL,
			HTTPClient: &http.Client{
				Timeout: 10 * time.Second,
			},
			Logger: logger,
		},
	}
}

type NoteService interface {
	GetNotesByCategory(ctx context.Context, categoryUuid, userUuid string) ([]byte, error)
	GetCalendarNotes(ctx context.Context, from, to int64, userUuid string) ([]byte, error)
	GetNote(ctx context.Context, uuid, userUuid string) ([]byte, error)
	GetStats(ctx context.Context, userUuid string) (NoteStats, error)
	CreateNote(ctx context.Context, dto CreateNoteDTO) (string, error)
	UpdateNote(ctx context.Context, uuid, userUuid string, dto UpdateNoteDTO) error
	DeleteNote(ctx context.Context, uuid, userUuid string) error
	GetTags(ctx context.Context, tagUUIDs []string, userUuid string) ([]byte, error)
	CreateTag(ctx context.Context, dto CreateTagDTO) (string, error)
	DeleteTag(ctx context.Context, uuid, userUuid string) error
}

func (c *client) GetNotesByCategory(ctx context.Context, categoryUuid, userUuid string) ([]byte, error) {
	var notes []byte

	filters := []rest.FilterOptions{
		{
			Field:  "category_uuid",
			Values: []string{categoryUuid},
		},
		{
			Field:  "user_uuid",
			Values: []string{userUuid},
		},
	}

	uri, err := c.base.BuildURL(c.Resource, filters)
	if err != nil {
		return notes, fmt.Errorf("failed to build URL. error: %w", err)
	}

	req, err := http.NewRequest(http.MethodGet, uri, nil)
	if err != nil {
		return notes, fmt.Errorf("failed to create new request due to error: %w", err)
	}

	reqCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	req = req.WithContext(reqCtx)

	response, err := c.base.SendRequest(req)
	if err != nil {
		return notes, fmt.Errorf("failed to send request due to error: %w", err)
	}

	if response.IsOk {
		notes, err = response.ReadBody()
		if err != nil {
			return nil, fmt.Errorf("failed to read body: %w", err)
		}
		return notes, nil
	}

	return nil, apperror.APIError(
		response.StatusCode(),
		response.Error.ErrorCode,
		response.Error.Message,
		response.Error.DeveloperMessage,
	)
}

func (c *client) GetNote(ctx context.Context, uuid, userUuid string) ([]byte, error) {
	var note []byte

	uri, err := c.base.BuildURL(fmt.Sprintf("%s/%s", c.Resource, uuid), []rest.FilterOptions{
		{
			Field:  "user_uuid",
			Values: []string{userUuid},
		},
	})
	if err != nil {
		return note, fmt.Errorf("failed to build URL. error: %w", err)
	}

	req, err := http.NewRequest(http.MethodGet, uri, nil)
	if err != nil {
		return note, fmt.Errorf("failed to create new request due to error: %w", err)
	}

	reqCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	req = req.WithContext(reqCtx)

	response, err := c.base.SendRequest(req)
	if err != nil {
		return note, fmt.Errorf("failed to send request due to error: %w", err)
	}

	if response.IsOk {
		note, err = response.ReadBody()
		if err != nil {
			return nil, fmt.Errorf("failed to read body: %w", err)
		}
		return note, nil
	}

	return nil, apperror.APIError(
		response.StatusCode(),
		response.Error.ErrorCode,
		response.Error.Message,
		response.Error.DeveloperMessage,
	)
}

func (c *client) GetCalendarNotes(ctx context.Context, from, to int64, userUuid string) ([]byte, error) {
	var notes []byte

	filters := []rest.FilterOptions{
		{
			Field:  "user_uuid",
			Values: []string{userUuid},
		},
		{
			Field:  "from",
			Values: []string{fmt.Sprintf("%d", from)},
		},
		{
			Field:  "to",
			Values: []string{fmt.Sprintf("%d", to)},
		},
	}

	uri, err := c.base.BuildURL("calendar", filters)
	if err != nil {
		return notes, fmt.Errorf("failed to build URL. error: %w", err)
	}

	req, err := http.NewRequest(http.MethodGet, uri, nil)
	if err != nil {
		return notes, fmt.Errorf("failed to create new request due to error: %w", err)
	}

	reqCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	req = req.WithContext(reqCtx)

	response, err := c.base.SendRequest(req)
	if err != nil {
		return notes, fmt.Errorf("failed to send request due to error: %w", err)
	}

	if response.IsOk {
		notes, err = response.ReadBody()
		if err != nil {
			return nil, fmt.Errorf("failed to read body: %w", err)
		}
		return notes, nil
	}

	return nil, apperror.APIError(
		response.StatusCode(),
		response.Error.ErrorCode,
		response.Error.Message,
		response.Error.DeveloperMessage,
	)
}

func (c *client) GetStats(ctx context.Context, userUuid string) (stats NoteStats, err error) {
	uri, err := c.base.BuildURL("stats", []rest.FilterOptions{
		{
			Field:  "user_uuid",
			Values: []string{userUuid},
		},
	})
	if err != nil {
		return stats, fmt.Errorf("failed to build URL. error: %w", err)
	}

	req, err := http.NewRequest(http.MethodGet, uri, nil)
	if err != nil {
		return stats, fmt.Errorf("failed to create new request due to error: %w", err)
	}

	reqCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	req = req.WithContext(reqCtx)

	response, err := c.base.SendRequest(req)
	if err != nil {
		return stats, fmt.Errorf("failed to send request due to error: %w", err)
	}

	if response.IsOk {
		body, err := response.ReadBody()
		if err != nil {
			return stats, fmt.Errorf("failed to read body: %w", err)
		}
		if err = json.Unmarshal(body, &stats); err != nil {
			return stats, fmt.Errorf("failed to unmarshal note stats: %w", err)
		}
		return stats, nil
	}

	return stats, apperror.APIError(
		response.StatusCode(),
		response.Error.ErrorCode,
		response.Error.Message,
		response.Error.DeveloperMessage,
	)
}

func (c *client) CreateNote(ctx context.Context, dto CreateNoteDTO) (string, error) {
	var noteUuid string

	uri, err := c.base.BuildURL(c.Resource, nil)
	if err != nil {
		return noteUuid, fmt.Errorf("failed to build URL. error: %w", err)
	}

	dataBytes, err := json.Marshal(dto)
	if err != nil {
		return noteUuid, fmt.Errorf("failed to marshal dto: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, uri, bytes.NewBuffer(dataBytes))
	if err != nil {
		return noteUuid, fmt.Errorf("failed to create new request due to error: %w", err)
	}

	reqCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	req = req.WithContext(reqCtx)

	response, err := c.base.SendRequest(req)
	if err != nil {
		return noteUuid, fmt.Errorf("failed to send request due to error: %w", err)
	}

	if response.IsOk {
		noteURL, err := response.Location()
		if err != nil {
			return noteUuid, fmt.Errorf("failed to get Location header: %w", err)
		}

		splitNoteURL := strings.Split(noteURL.String(), "/")
		noteUuid = splitNoteURL[len(splitNoteURL)-1]
		return noteUuid, nil
	}

	return noteUuid, apperror.APIError(
		response.StatusCode(),
		response.Error.ErrorCode,
		response.Error.Message,
		response.Error.DeveloperMessage,
	)
}

func (c *client) UpdateNote(ctx context.Context, uuid, userUuid string, dto UpdateNoteDTO) error {
	uri, err := c.base.BuildURL(fmt.Sprintf("%s/%s", c.Resource, uuid), []rest.FilterOptions{
		{
			Field:  "user_uuid",
			Values: []string{userUuid},
		},
	})
	if err != nil {
		return fmt.Errorf("failed to build URL. error: %w", err)
	}

	dataBytes, err := json.Marshal(dto)
	if err != nil {
		return fmt.Errorf("failed to marshal dto: %w", err)
	}

	req, err := http.NewRequest(http.MethodPatch, uri, bytes.NewBuffer(dataBytes))
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

func (c *client) DeleteNote(ctx context.Context, uuid, userUuid string) error {
	uri, err := c.base.BuildURL(fmt.Sprintf("%s/%s", c.Resource, uuid), []rest.FilterOptions{
		{
			Field:  "user_uuid",
			Values: []string{userUuid},
		},
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

func (c *client) GetTags(ctx context.Context, tagUUIDs []string, userUuid string) ([]byte, error) {
	var tags []byte

	filters := []rest.FilterOptions{
		{
			Field:  "user_uuid",
			Values: []string{userUuid},
		},
	}
	if len(tagUUIDs) > 0 {
		filters = append(filters,
			rest.FilterOptions{
				Field:  "id",
				Values: tagUUIDs,
			},
		)
	}

	uri, err := c.base.BuildURL("tags", filters)
	if err != nil {
		return tags, fmt.Errorf("failed to build URL. error: %w", err)
	}

	req, err := http.NewRequest(http.MethodGet, uri, nil)
	if err != nil {
		return tags, fmt.Errorf("failed to create new request due to error: %w", err)
	}

	reqCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	req = req.WithContext(reqCtx)

	response, err := c.base.SendRequest(req)
	if err != nil {
		return tags, fmt.Errorf("failed to send request due to error: %w", err)
	}

	if response.IsOk {
		tags, err = response.ReadBody()
		if err != nil {
			return nil, fmt.Errorf("failed to read body: %w", err)
		}
		return tags, nil
	}

	return nil, apperror.APIError(
		response.StatusCode(),
		response.Error.ErrorCode,
		response.Error.Message,
		response.Error.DeveloperMessage,
	)
}

func (c *client) CreateTag(ctx context.Context, dto CreateTagDTO) (string, error) {
	var tagUuid string

	uri, err := c.base.BuildURL("tags", nil)
	if err != nil {
		return tagUuid, fmt.Errorf("failed to build URL. error: %w", err)
	}

	dataBytes, err := json.Marshal(dto)
	if err != nil {
		return tagUuid, fmt.Errorf("failed to marshal dto: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, uri, bytes.NewBuffer(dataBytes))
	if err != nil {
		return tagUuid, fmt.Errorf("failed to create new request due to error: %w", err)
	}

	reqCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	req = req.WithContext(reqCtx)

	response, err := c.base.SendRequest(req)
	if err != nil {
		return tagUuid, fmt.Errorf("failed to send request due to error: %w", err)
	}

	if response.IsOk {
		tagURL, err := response.Location()
		if err != nil {
			return tagUuid, fmt.Errorf("failed to get Location header: %w", err)
		}

		splitTagURL := strings.Split(tagURL.String(), "/")
		tagUuid = splitTagURL[len(splitTagURL)-1]
		return tagUuid, nil
	}

	return tagUuid, apperror.APIError(
		response.StatusCode(),
		response.Error.ErrorCode,
		response.Error.Message,
		response.Error.DeveloperMessage,
	)
}

func (c *client) DeleteTag(ctx context.Context, uuid, userUuid string) error {
	uri, err := c.base.BuildURL(fmt.Sprintf("tags/%s", uuid), []rest.FilterOptions{
		{
			Field:  "user_uuid",
			Values: []string{userUuid},
		},
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
