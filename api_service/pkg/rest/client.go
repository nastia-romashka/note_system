package rest

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"

	"myproject/pkg/logging"
)

type BaseClient struct {
	BaseURL    string
	HTTPClient *http.Client
	Logger     logging.Logger
}

type APIError struct {
	ErrorCode        string `json:"code"`
	Message          string `json:"message"`
	DeveloperMessage string `json:"developer_message"`
}

type APIResponse struct {
	IsOk     bool
	Error    APIError
	response *http.Response
}

func (r *APIResponse) ReadBody() ([]byte, error) {
	if r == nil || r.response == nil || r.response.Body == nil {
		return nil, errors.New("response body is empty")
	}
	defer r.response.Body.Close()

	body, err := io.ReadAll(r.response.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return body, nil
}

func (r *APIResponse) Location() (*url.URL, error) {
	if r == nil || r.response == nil {
		return nil, errors.New("response is empty")
	}

	location := r.response.Header.Get("Location")
	if location == "" {
		return nil, errors.New("location header is empty")
	}

	locationURL, err := url.Parse(location)
	if err != nil {
		return nil, fmt.Errorf("failed to parse location header: %w", err)
	}

	return locationURL, nil
}

func (r *APIResponse) StatusCode() int {
	if r == nil || r.response == nil {
		return 0
	}

	return r.response.StatusCode
}

func (c *BaseClient) SendRequest(req *http.Request) (*APIResponse, error) {
	if c.HTTPClient == nil {
		return nil, errors.New("no http client")
	}

	req.Header.Set("Accept", "application/json; charset=utf-8")
	req.Header.Set("Content-Type", "application/json; charset=utf-8")

	response, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request. error: %w", err)
	}

	apiResponse := APIResponse{
		IsOk:     true,
		response: response,
	}
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusBadRequest {
		apiResponse.IsOk = false
		defer response.Body.Close()

		var apiErr APIError
		if err = json.NewDecoder(response.Body).Decode(&apiErr); err == nil {
			apiResponse.Error = apiErr
		}
	}

	return &apiResponse, nil
}

func (c *BaseClient) BuildURL(resource string, filters []FilterOptions) (string, error) {
	var resultURL string

	parsedURL, err := url.ParseRequestURI(c.BaseURL)
	if err != nil {
		return resultURL, fmt.Errorf("failed to parse base URL. error: %w", err)
	}
	parsedURL.Path = path.Join(parsedURL.Path, resource)

	if len(filters) > 0 {
		q := parsedURL.Query()
		for _, fo := range filters {
			q.Set(fo.Field, fo.ToStringWF())
		}
		parsedURL.RawQuery = q.Encode()
	}

	return parsedURL.String(), nil
}

func (c *BaseClient) Close() error {
	c.HTTPClient = nil
	return nil
}
