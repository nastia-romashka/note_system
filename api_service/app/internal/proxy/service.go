package proxy

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"time"
)

type Service struct {
	BaseURL    string
	HTTPClient *http.Client
}

func NewService(baseURL string, timeout time.Duration) Service {
	return Service{
		BaseURL: baseURL,
		HTTPClient: &http.Client{
			Timeout: timeout,
		},
	}
}

func (s Service) Forward(
	w http.ResponseWriter,
	r *http.Request,
	method string,
	resource string,
	query url.Values,
	body io.Reader,
	headers http.Header,
) error {
	targetURL, err := s.buildURL(resource, query)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(r.Context(), method, targetURL, body)
	if err != nil {
		return fmt.Errorf("create proxy request: %w", err)
	}

	copyHeaders(req.Header, r.Header)
	copyHeaders(req.Header, headers)

	response, err := s.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("send proxy request: %w", err)
	}
	defer response.Body.Close()

	copyHeaders(w.Header(), response.Header)
	w.WriteHeader(response.StatusCode)
	if _, err = io.Copy(w, response.Body); err != nil {
		return fmt.Errorf("copy proxy response body: %w", err)
	}

	return nil
}

func (s Service) buildURL(resource string, query url.Values) (string, error) {
	parsedURL, err := url.ParseRequestURI(s.BaseURL)
	if err != nil {
		return "", fmt.Errorf("parse proxy base URL: %w", err)
	}

	parsedURL.Path = path.Join(parsedURL.Path, resource)
	if len(query) > 0 {
		parsedURL.RawQuery = query.Encode()
	}

	return parsedURL.String(), nil
}

func copyHeaders(dst, src http.Header) {
	for key, values := range src {
		for _, value := range values {
			dst.Add(key, value)
		}
	}
}
