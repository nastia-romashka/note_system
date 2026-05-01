package user

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

func NewService(baseURL string, resource string, logger logging.Logger) UserService {
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

type UserService interface {
	CreateUser(ctx context.Context, dto CreateUserDTO) (string, error)
	GetUser(ctx context.Context, uuid string) (User, error)
	Authenticate(ctx context.Context, dto AuthUserDTO) (User, error)
}

func (c *client) CreateUser(ctx context.Context, dto CreateUserDTO) (string, error) {
	var userUuid string

	uri, err := c.base.BuildURL(c.Resource, nil)
	if err != nil {
		return userUuid, fmt.Errorf("failed to build URL. error: %w", err)
	}

	dataBytes, err := json.Marshal(dto)
	if err != nil {
		return userUuid, fmt.Errorf("failed to marshal dto: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, uri, bytes.NewBuffer(dataBytes))
	if err != nil {
		return userUuid, fmt.Errorf("failed to create new request due to error: %w", err)
	}

	reqCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	req = req.WithContext(reqCtx)

	response, err := c.base.SendRequest(req)
	if err != nil {
		return userUuid, fmt.Errorf("failed to send request due to error: %w", err)
	}

	if response.IsOk {
		userURL, err := response.Location()
		if err != nil {
			return userUuid, fmt.Errorf("failed to get Location header: %w", err)
		}

		splitUserURL := strings.Split(userURL.String(), "/")
		userUuid = splitUserURL[len(splitUserURL)-1]
		return userUuid, nil
	}

	return userUuid, apperror.APIError(
		response.StatusCode(),
		response.Error.ErrorCode,
		response.Error.Message,
		response.Error.DeveloperMessage,
	)
}

func (c *client) GetUser(ctx context.Context, uuid string) (user User, err error) {
	uri, err := c.base.BuildURL(fmt.Sprintf("%s/%s", c.Resource, uuid), nil)
	if err != nil {
		return user, fmt.Errorf("failed to build URL. error: %w", err)
	}

	req, err := http.NewRequest(http.MethodGet, uri, nil)
	if err != nil {
		return user, fmt.Errorf("failed to create new request due to error: %w", err)
	}

	reqCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	req = req.WithContext(reqCtx)

	response, err := c.base.SendRequest(req)
	if err != nil {
		return user, fmt.Errorf("failed to send request due to error: %w", err)
	}

	if response.IsOk {
		body, err := response.ReadBody()
		if err != nil {
			return user, fmt.Errorf("failed to read body: %w", err)
		}

		if err = json.Unmarshal(body, &user); err != nil {
			return user, fmt.Errorf("failed to unmarshal user: %w", err)
		}
		return user, nil
	}

	return user, apperror.APIError(
		response.StatusCode(),
		response.Error.ErrorCode,
		response.Error.Message,
		response.Error.DeveloperMessage,
	)
}

func (c *client) Authenticate(ctx context.Context, dto AuthUserDTO) (user User, err error) {
	uri, err := c.base.BuildURL(fmt.Sprintf("%s/authenticate", c.Resource), nil)
	if err != nil {
		return user, fmt.Errorf("failed to build URL. error: %w", err)
	}

	dataBytes, err := json.Marshal(dto)
	if err != nil {
		return user, fmt.Errorf("failed to marshal dto: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, uri, bytes.NewBuffer(dataBytes))
	if err != nil {
		return user, fmt.Errorf("failed to create new request due to error: %w", err)
	}

	reqCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	req = req.WithContext(reqCtx)

	response, err := c.base.SendRequest(req)
	if err != nil {
		return user, fmt.Errorf("failed to send request due to error: %w", err)
	}

	if response.IsOk {
		body, err := response.ReadBody()
		if err != nil {
			return user, fmt.Errorf("failed to read body: %w", err)
		}

		if err = json.Unmarshal(body, &user); err != nil {
			return user, fmt.Errorf("failed to unmarshal user: %w", err)
		}
		return user, nil
	}

	return user, apperror.APIError(
		response.StatusCode(),
		response.Error.ErrorCode,
		response.Error.Message,
		response.Error.DeveloperMessage,
	)
}
