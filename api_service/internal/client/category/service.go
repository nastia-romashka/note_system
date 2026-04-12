package category

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

func NewService(baseURL string, resource string, logger logging.Logger) CategoryService {
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

type CategoryService interface {
	GetUserCategories(ctx context.Context, userUuid string) ([]byte, error)
	CreateCategory(ctx context.Context, dto CreateCategoryDTO) (string, error)
	UpdateCategory(ctx context.Context, uuid string, dto UpdateCategoryDTO) error
	DeleteCategory(ctx context.Context, dto DeleteCategoryDTO) error
}

func (c *client) GetUserCategories(ctx context.Context, userUuid string) ([]byte, error) {
	var categories []byte

	c.base.Logger.Debug("add user_uuid to filter options")
	filters := []rest.FilterOptions{
		{
			Field:  "user_uuid",
			Values: []string{userUuid},
		},
	}

	c.base.Logger.Debug("build url with resource and filter")
	uri, err := c.base.BuildURL(c.Resource, filters)
	if err != nil {
		return categories, fmt.Errorf("failed to build URL. error: %w", err)
	}
	c.base.Logger.Tracef("url: %s", uri)

	c.base.Logger.Debug("create new request")
	req, err := http.NewRequest(http.MethodGet, uri, nil)
	if err != nil {
		return categories, fmt.Errorf("failed to create new request due to error: %w", err)
	}

	c.base.Logger.Debug("send request")
	reqCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	req = req.WithContext(reqCtx)
	response, err := c.base.SendRequest(req)
	if err != nil {
		return categories, fmt.Errorf("failed to send request due to error: %w", err)
	}

	if response.IsOk {
		c.base.Logger.Debug("read body")
		categories, err = response.ReadBody()
		if err != nil {
			return nil, fmt.Errorf("failed to read body: %w", err)
		}
		return categories, nil
	}

	return nil, apperror.APIError(
		response.StatusCode(),
		response.Error.ErrorCode,
		response.Error.Message,
		response.Error.DeveloperMessage,
	)
}

func (c *client) CreateCategory(ctx context.Context, dto CreateCategoryDTO) (string, error) {
	var categoryUuid string

	c.base.Logger.Debug("build url with resource")
	uri, err := c.base.BuildURL(c.Resource, nil)
	if err != nil {
		return categoryUuid, fmt.Errorf("failed to build URL. error: %w", err)
	}

	c.base.Logger.Debug("marshal dto to bytes")
	dataBytes, err := json.Marshal(dto)
	if err != nil {
		return categoryUuid, fmt.Errorf("failed to marshal dto: %w", err)
	}

	c.base.Logger.Debug("create new request")
	req, err := http.NewRequest(http.MethodPost, uri, bytes.NewBuffer(dataBytes))
	if err != nil {
		return categoryUuid, fmt.Errorf("failed to create new request due to error: %w", err)
	}

	c.base.Logger.Debug("send request")
	reqCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	req = req.WithContext(reqCtx)
	response, err := c.base.SendRequest(req)
	if err != nil {
		return categoryUuid, fmt.Errorf("failed to send request due to error: %w", err)
	}

	if response.IsOk {
		c.base.Logger.Debug("parse location header")
		categoryURL, err := response.Location()
		if err != nil {
			return categoryUuid, fmt.Errorf("failed to get Location header: %w", err)
		}
		c.base.Logger.Tracef("Location: %s", categoryURL.String())

		splitCategoryURL := strings.Split(categoryURL.String(), "/")
		categoryUuid = splitCategoryURL[len(splitCategoryURL)-1]
		return categoryUuid, nil
	}

	return categoryUuid, apperror.APIError(
		response.StatusCode(),
		response.Error.ErrorCode,
		response.Error.Message,
		response.Error.DeveloperMessage,
	)
}

func (c *client) UpdateCategory(ctx context.Context, uuid string, dto UpdateCategoryDTO) error {
	c.base.Logger.Debug("build url with resource and category uuid")
	uri, err := c.base.BuildURL(fmt.Sprintf("%s/%s", c.Resource, uuid), nil)
	if err != nil {
		return fmt.Errorf("failed to build URL. error: %w", err)
	}

	c.base.Logger.Debug("marshal dto to bytes")
	dataBytes, err := json.Marshal(dto)
	if err != nil {
		return fmt.Errorf("failed to marshal dto: %w", err)
	}

	c.base.Logger.Debug("create new request")
	req, err := http.NewRequest(http.MethodPatch, uri, bytes.NewBuffer(dataBytes))
	if err != nil {
		return fmt.Errorf("failed to create new request due to error: %w", err)
	}

	c.base.Logger.Debug("send request")
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

func (c *client) DeleteCategory(ctx context.Context, dto DeleteCategoryDTO) error {
	c.base.Logger.Debug("build url with resource and category uuid")
	uri, err := c.base.BuildURL(fmt.Sprintf("%s/%s", c.Resource, dto.Uuid), nil)
	if err != nil {
		return fmt.Errorf("failed to build URL. error: %w", err)
	}

	c.base.Logger.Debug("marshal dto to bytes")
	dataBytes, err := json.Marshal(dto)
	if err != nil {
		return fmt.Errorf("failed to marshal dto: %w", err)
	}

	c.base.Logger.Debug("create new request")
	req, err := http.NewRequest(http.MethodDelete, uri, bytes.NewBuffer(dataBytes))
	if err != nil {
		return fmt.Errorf("failed to create new request due to error: %w", err)
	}

	c.base.Logger.Debug("send request")
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
