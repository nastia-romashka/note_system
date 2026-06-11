package user

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"path"
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
				Timeout: 30 * time.Second,
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
	GetWorkspace(ctx context.Context, workspaceUUID string) (Workspace, error)
	GetWorkspaceMembers(ctx context.Context, workspaceUUID string) ([]WorkspaceMember, error)
	UpdateWorkspaceMember(ctx context.Context, workspaceUUID, memberUserUUID string, dto UpdateWorkspaceMemberDTO) (WorkspaceMember, error)
	GetWorkspaceInvites(ctx context.Context, workspaceUUID, userUUID string) ([]WorkspaceInvite, error)
	CreateWorkspaceInvite(ctx context.Context, workspaceUUID string, dto CreateWorkspaceInviteDTO) (WorkspaceInvite, error)
	GetPersonalWorkspace(ctx context.Context, userUUID string) (Workspace, error)
	GetWorkspaceAccess(ctx context.Context, workspaceUUID, userUUID string) (WorkspaceAccess, error)
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

	reqCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
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

	reqCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
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

	reqCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
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

	reqCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
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

	reqCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
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

	reqCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
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

	reqCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
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

	reqCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
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

	reqCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
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

	reqCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
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

func (c *client) GetWorkspace(ctx context.Context, workspaceUUID string) (workspace Workspace, err error) {
	uri, err := c.base.BuildURL(fmt.Sprintf("workspaces/%s", workspaceUUID), nil)
	if err != nil {
		return workspace, fmt.Errorf("failed to build URL. error: %w", err)
	}

	req, err := http.NewRequest(http.MethodGet, uri, nil)
	if err != nil {
		return workspace, fmt.Errorf("failed to create new request due to error: %w", err)
	}

	reqCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	req = req.WithContext(reqCtx)

	response, err := c.base.SendRequest(req)
	if err != nil {
		return workspace, fmt.Errorf("failed to send request due to error: %w", err)
	}

	if response.IsOk {
		body, err := response.ReadBody()
		if err != nil {
			return workspace, fmt.Errorf("failed to read body: %w", err)
		}

		if err = json.Unmarshal(body, &workspace); err != nil {
			return workspace, fmt.Errorf("failed to unmarshal workspace: %w", err)
		}
		return workspace, nil
	}

	return workspace, apperror.APIError(
		response.StatusCode(),
		response.Error.ErrorCode,
		response.Error.Message,
		response.Error.DeveloperMessage,
	)
}

func (c *client) GetWorkspaceMembers(ctx context.Context, workspaceUUID string) (members []WorkspaceMember, err error) {
	uri, err := c.base.BuildURL(fmt.Sprintf("workspaces/%s/members", workspaceUUID), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to build URL. error: %w", err)
	}

	req, err := http.NewRequest(http.MethodGet, uri, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create new request due to error: %w", err)
	}

	reqCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
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

		if err = json.Unmarshal(body, &members); err != nil {
			return nil, fmt.Errorf("failed to unmarshal workspace members: %w", err)
		}
		return members, nil
	}

	return nil, apperror.APIError(
		response.StatusCode(),
		response.Error.ErrorCode,
		response.Error.Message,
		response.Error.DeveloperMessage,
	)
}

func (c *client) UpdateWorkspaceMember(
	ctx context.Context,
	workspaceUUID,
	memberUserUUID string,
	dto UpdateWorkspaceMemberDTO,
) (member WorkspaceMember, err error) {
	uri, err := c.base.BuildURL(fmt.Sprintf("workspaces/%s/members/%s", workspaceUUID, memberUserUUID), nil)
	if err != nil {
		return member, fmt.Errorf("failed to build URL. error: %w", err)
	}

	dataBytes, err := json.Marshal(dto)
	if err != nil {
		return member, fmt.Errorf("failed to marshal dto: %w", err)
	}

	req, err := http.NewRequest(http.MethodPatch, uri, bytes.NewBuffer(dataBytes))
	if err != nil {
		return member, fmt.Errorf("failed to create new request due to error: %w", err)
	}

	reqCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	req = req.WithContext(reqCtx)

	response, err := c.base.SendRequest(req)
	if err != nil {
		return member, fmt.Errorf("failed to send request due to error: %w", err)
	}

	if response.IsOk {
		body, err := response.ReadBody()
		if err != nil {
			return member, fmt.Errorf("failed to read body: %w", err)
		}

		if err = json.Unmarshal(body, &member); err != nil {
			return member, fmt.Errorf("failed to unmarshal updated workspace member: %w", err)
		}
		return member, nil
	}

	return member, apperror.APIError(
		response.StatusCode(),
		response.Error.ErrorCode,
		response.Error.Message,
		response.Error.DeveloperMessage,
	)
}

func (c *client) GetWorkspaceInvites(ctx context.Context, workspaceUUID, userUUID string) (invites []WorkspaceInvite, err error) {
	uri, err := c.base.BuildURL(fmt.Sprintf("workspaces/%s/invites", workspaceUUID), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to build URL. error: %w", err)
	}

	parsedURL, err := url.Parse(uri)
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL. error: %w", err)
	}
	query := parsedURL.Query()
	query.Set("user_id", userUUID)
	parsedURL.RawQuery = query.Encode()

	req, err := http.NewRequest(http.MethodGet, parsedURL.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create new request due to error: %w", err)
	}

	reqCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
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

		if err = json.Unmarshal(body, &invites); err != nil {
			return nil, fmt.Errorf("failed to unmarshal workspace invites: %w", err)
		}
		return invites, nil
	}

	return nil, apperror.APIError(
		response.StatusCode(),
		response.Error.ErrorCode,
		response.Error.Message,
		response.Error.DeveloperMessage,
	)
}

func (c *client) CreateWorkspaceInvite(ctx context.Context, workspaceUUID string, dto CreateWorkspaceInviteDTO) (invite WorkspaceInvite, err error) {
	uri, err := c.base.BuildURL(fmt.Sprintf("workspaces/%s/invites", workspaceUUID), nil)
	if err != nil {
		return invite, fmt.Errorf("failed to build URL. error: %w", err)
	}

	dataBytes, err := json.Marshal(dto)
	if err != nil {
		return invite, fmt.Errorf("failed to marshal dto: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, uri, bytes.NewBuffer(dataBytes))
	if err != nil {
		return invite, fmt.Errorf("failed to create new request due to error: %w", err)
	}

	reqCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	req = req.WithContext(reqCtx)

	response, err := c.base.SendRequest(req)
	if err != nil {
		return invite, fmt.Errorf("failed to send request due to error: %w", err)
	}

	if response.IsOk {
		body, err := response.ReadBody()
		if err != nil {
			return invite, fmt.Errorf("failed to read body: %w", err)
		}

		if err = json.Unmarshal(body, &invite); err != nil {
			return invite, fmt.Errorf("failed to unmarshal workspace invite: %w", err)
		}
		return invite, nil
	}

	return invite, apperror.APIError(
		response.StatusCode(),
		response.Error.ErrorCode,
		response.Error.Message,
		response.Error.DeveloperMessage,
	)
}

func (c *client) GetPersonalWorkspace(ctx context.Context, userUUID string) (workspace Workspace, err error) {
	uri, err := c.buildInternalURL(fmt.Sprintf("internal/users/%s/personal-workspace", userUUID))
	if err != nil {
		return workspace, fmt.Errorf("failed to build URL. error: %w", err)
	}

	req, err := http.NewRequest(http.MethodGet, uri, nil)
	if err != nil {
		return workspace, fmt.Errorf("failed to create new request due to error: %w", err)
	}

	reqCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	req = req.WithContext(reqCtx)

	response, err := c.base.SendRequest(req)
	if err != nil {
		return workspace, fmt.Errorf("failed to send request due to error: %w", err)
	}

	if response.IsOk {
		body, err := response.ReadBody()
		if err != nil {
			return workspace, fmt.Errorf("failed to read body: %w", err)
		}

		if err = json.Unmarshal(body, &workspace); err != nil {
			return workspace, fmt.Errorf("failed to unmarshal workspace: %w", err)
		}
		return workspace, nil
	}

	return workspace, apperror.APIError(
		response.StatusCode(),
		response.Error.ErrorCode,
		response.Error.Message,
		response.Error.DeveloperMessage,
	)
}

func (c *client) GetWorkspaceAccess(ctx context.Context, workspaceUUID, userUUID string) (access WorkspaceAccess, err error) {
	uri, err := c.buildInternalURL(fmt.Sprintf("internal/workspaces/%s/access", workspaceUUID))
	if err != nil {
		return access, fmt.Errorf("failed to build URL. error: %w", err)
	}

	parsedURL, err := url.Parse(uri)
	if err != nil {
		return access, fmt.Errorf("failed to parse URL. error: %w", err)
	}
	query := parsedURL.Query()
	query.Set("user_id", userUUID)
	parsedURL.RawQuery = query.Encode()

	req, err := http.NewRequest(http.MethodGet, parsedURL.String(), nil)
	if err != nil {
		return access, fmt.Errorf("failed to create new request due to error: %w", err)
	}

	reqCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	req = req.WithContext(reqCtx)

	response, err := c.base.SendRequest(req)
	if err != nil {
		return access, fmt.Errorf("failed to send request due to error: %w", err)
	}

	if response.IsOk {
		body, err := response.ReadBody()
		if err != nil {
			return access, fmt.Errorf("failed to read body: %w", err)
		}

		if err = json.Unmarshal(body, &access); err != nil {
			return access, fmt.Errorf("failed to unmarshal workspace access: %w", err)
		}
		return access, nil
	}

	return access, apperror.APIError(
		response.StatusCode(),
		response.Error.ErrorCode,
		response.Error.Message,
		response.Error.DeveloperMessage,
	)
}

func (c *client) buildInternalURL(resource string) (string, error) {
	parsedURL, err := url.ParseRequestURI(c.base.BaseURL)
	if err != nil {
		return "", fmt.Errorf("failed to parse base URL. error: %w", err)
	}

	basePath := strings.TrimSuffix(strings.TrimSuffix(parsedURL.Path, "/"), "/api")
	if basePath == "" {
		basePath = "/"
	}

	parsedURL.Path = path.Join(basePath, resource)
	return parsedURL.String(), nil
}
