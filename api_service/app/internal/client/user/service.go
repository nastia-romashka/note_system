package user

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
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
	GetProfile(ctx context.Context, uuid string) (UserProfile, error)
	UpdateProfile(ctx context.Context, uuid string, dto UpdateUserProfileDTO) error
	GetActions(ctx context.Context, uuid string, limit, offset int) ([]UserAction, error)
	CreateAction(ctx context.Context, uuid string, dto CreateUserActionDTO) error
	Authenticate(ctx context.Context, dto AuthUserDTO) (User, error)
	CreateSession(ctx context.Context, dto CreateUserSessionDTO) error
	RotateSession(ctx context.Context, dto RotateUserSessionDTO) (UserSession, error)
	RevokeSession(ctx context.Context, refreshTokenHash string) error
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

func (c *client) GetProfile(ctx context.Context, uuid string) (profile UserProfile, err error) {
	uri, err := c.base.BuildURL(fmt.Sprintf("%s/%s/profile", c.Resource, uuid), nil)
	if err != nil {
		return profile, fmt.Errorf("failed to build URL. error: %w", err)
	}

	req, err := http.NewRequest(http.MethodGet, uri, nil)
	if err != nil {
		return profile, fmt.Errorf("failed to create new request due to error: %w", err)
	}

	reqCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	req = req.WithContext(reqCtx)

	response, err := c.base.SendRequest(req)
	if err != nil {
		return profile, fmt.Errorf("failed to send request due to error: %w", err)
	}

	if response.IsOk {
		body, err := response.ReadBody()
		if err != nil {
			return profile, fmt.Errorf("failed to read body: %w", err)
		}

		if err = json.Unmarshal(body, &profile); err != nil {
			return profile, fmt.Errorf("failed to unmarshal user profile: %w", err)
		}
		return profile, nil
	}

	return profile, apperror.APIError(
		response.StatusCode(),
		response.Error.ErrorCode,
		response.Error.Message,
		response.Error.DeveloperMessage,
	)
}

func (c *client) UpdateProfile(ctx context.Context, uuid string, dto UpdateUserProfileDTO) error {
	uri, err := c.base.BuildURL(fmt.Sprintf("%s/%s/profile", c.Resource, uuid), nil)
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

func (c *client) GetActions(ctx context.Context, uuid string, limit, offset int) (actions []UserAction, err error) {
	uri, err := c.base.BuildURL(fmt.Sprintf("%s/%s/actions", c.Resource, uuid), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to build URL. error: %w", err)
	}

	parsedURL, err := url.Parse(uri)
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL. error: %w", err)
	}
	query := parsedURL.Query()
	query.Set("limit", strconv.Itoa(limit))
	query.Set("offset", strconv.Itoa(offset))
	parsedURL.RawQuery = query.Encode()

	req, err := http.NewRequest(http.MethodGet, parsedURL.String(), nil)
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

		if err = json.Unmarshal(body, &actions); err != nil {
			return nil, fmt.Errorf("failed to unmarshal user actions: %w", err)
		}
		return actions, nil
	}

	return nil, apperror.APIError(
		response.StatusCode(),
		response.Error.ErrorCode,
		response.Error.Message,
		response.Error.DeveloperMessage,
	)
}

func (c *client) CreateAction(ctx context.Context, uuid string, dto CreateUserActionDTO) (err error) {
	uri, err := c.base.BuildURL(fmt.Sprintf("user-actions/%s", uuid), nil)
	if err != nil {
		return fmt.Errorf("failed to build URL. error: %w", err)
	}

	dataBytes, err := json.Marshal(dto)
	if err != nil {
		return fmt.Errorf("failed to marshal dto: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, uri, bytes.NewBuffer(dataBytes))
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
		if response.Body() != nil {
			_ = response.Body().Close()
		}
		return nil
	}

	return apperror.APIError(
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

func (c *client) CreateSession(ctx context.Context, dto CreateUserSessionDTO) error {
	uri, err := c.base.BuildURL("user-sessions", nil)
	if err != nil {
		return fmt.Errorf("failed to build URL. error: %w", err)
	}

	dataBytes, err := json.Marshal(dto)
	if err != nil {
		return fmt.Errorf("failed to marshal dto: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, uri, bytes.NewBuffer(dataBytes))
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
		if response.Body() != nil {
			_ = response.Body().Close()
		}
		return nil
	}

	return apperror.APIError(
		response.StatusCode(),
		response.Error.ErrorCode,
		response.Error.Message,
		response.Error.DeveloperMessage,
	)
}

func (c *client) RotateSession(ctx context.Context, dto RotateUserSessionDTO) (session UserSession, err error) {
	uri, err := c.base.BuildURL("user-sessions/rotate", nil)
	if err != nil {
		return session, fmt.Errorf("failed to build URL. error: %w", err)
	}

	dataBytes, err := json.Marshal(dto)
	if err != nil {
		return session, fmt.Errorf("failed to marshal dto: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, uri, bytes.NewBuffer(dataBytes))
	if err != nil {
		return session, fmt.Errorf("failed to create new request due to error: %w", err)
	}

	reqCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	req = req.WithContext(reqCtx)

	response, err := c.base.SendRequest(req)
	if err != nil {
		return session, fmt.Errorf("failed to send request due to error: %w", err)
	}

	if response.IsOk {
		body, err := response.ReadBody()
		if err != nil {
			return session, fmt.Errorf("failed to read body: %w", err)
		}

		if err = json.Unmarshal(body, &session); err != nil {
			return session, fmt.Errorf("failed to unmarshal user session: %w", err)
		}
		return session, nil
	}

	return session, apperror.APIError(
		response.StatusCode(),
		response.Error.ErrorCode,
		response.Error.Message,
		response.Error.DeveloperMessage,
	)
}

func (c *client) RevokeSession(ctx context.Context, refreshTokenHash string) error {
	uri, err := c.base.BuildURL("user-sessions/revoke", nil)
	if err != nil {
		return fmt.Errorf("failed to build URL. error: %w", err)
	}

	dataBytes, err := json.Marshal(RevokeUserSessionDTO{RefreshTokenHash: refreshTokenHash})
	if err != nil {
		return fmt.Errorf("failed to marshal dto: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, uri, bytes.NewBuffer(dataBytes))
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
		if response.Body() != nil {
			_ = response.Body().Close()
		}
		return nil
	}

	return apperror.APIError(
		response.StatusCode(),
		response.Error.ErrorCode,
		response.Error.Message,
		response.Error.DeveloperMessage,
	)
}
