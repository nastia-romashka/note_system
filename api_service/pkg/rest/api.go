package rest

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
)

type APIResponse struct {
	IsOk     bool
	response *http.Response
	Error    APIError
}

func (ar *APIResponse) Body() io.ReadCloser {
	if ar == nil || ar.response == nil {
		return nil
	}

	return ar.response.Body
}

func (ar *APIResponse) Header() http.Header {
	if ar == nil || ar.response == nil {
		return nil
	}

	return ar.response.Header
}

func (ar *APIResponse) ReadBody() ([]byte, error) {
	if ar == nil || ar.response == nil || ar.response.Body == nil {
		return nil, fmt.Errorf("response body is empty")
	}

	defer ar.response.Body.Close()
	return io.ReadAll(ar.response.Body)
}

func (ar *APIResponse) StatusCode() int {
	if ar == nil || ar.response == nil {
		return 0
	}

	return ar.response.StatusCode
}

func (ar *APIResponse) Location() (*url.URL, error) {
	if ar == nil || ar.response == nil {
		return nil, fmt.Errorf("response is empty")
	}

	return ar.response.Location()
}

type APIError struct {
	Message          string `json:"message,omitempty"`
	ErrorCode        string `json:"code,omitempty"`
	DeveloperMessage string `json:"developer_message,omitempty"`
}

func (aep *APIError) ToString() string {
	return fmt.Sprintf(
		"Err Code: %s, Err: %s, Developer Err: %s",
		aep.ErrorCode,
		aep.Message,
		aep.DeveloperMessage,
	)
}
