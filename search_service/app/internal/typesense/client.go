package typesense

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	"search_service/internal/apperror"
	handlermodel "search_service/internal/handlers/notes"
	"search_service/pkg/logging"
)

type Client struct {
	baseURL    *url.URL
	apiKey     string
	collection string
	httpClient *http.Client
	logger     logging.Logger
}

func NewClient(baseURL, apiKey, collection string, logger logging.Logger) (*Client, error) {
	parsedURL, err := url.ParseRequestURI(strings.TrimSpace(baseURL))
	if err != nil {
		return nil, fmt.Errorf("parse typesense url: %w", err)
	}

	client := &Client{
		baseURL:    parsedURL,
		apiKey:     strings.TrimSpace(apiKey),
		collection: strings.TrimSpace(collection),
		httpClient: &http.Client{Timeout: 10 * time.Second},
		logger:     logger.With("layer", "typesense", "collection", collection),
	}

	if client.collection == "" {
		return nil, errors.New("typesense collection is empty")
	}

	if err := client.ensureCollection(context.Background()); err != nil {
		return nil, err
	}

	return client, nil
}

func (c *Client) Search(q, userUUID, categoryUUID string, tagUUIDs []string, page, perPage int) ([]handlermodel.SearchNote, error) {
	query := url.Values{}
	if strings.TrimSpace(q) == "" {
		query.Set("q", "*")
	} else {
		query.Set("q", q)
	}
	query.Set("query_by", "header,body,short_body,category_name,tag_names")
	query.Set("sort_by", "created_date:desc")
	query.Set("highlight_fields", "header,body,short_body")
	query.Set("page", fmt.Sprintf("%d", page))
	query.Set("per_page", fmt.Sprintf("%d", perPage))
	query.Set("filter_by", buildFilter(userUUID, categoryUUID, tagUUIDs))

	body, err := c.doRequest(context.Background(), http.MethodGet, path.Join("collections", c.collection, "documents", "search"), query, nil, http.StatusOK)
	if err != nil {
		return nil, err
	}

	var payload searchResponse
	if err = json.Unmarshal(body, &payload); err != nil {
		return nil, fmt.Errorf("decode search response: %w", err)
	}

	result := make([]handlermodel.SearchNote, 0, len(payload.Hits))
	for _, hit := range payload.Hits {
		result = append(result, handlermodel.SearchNote{
			Uuid:         hit.Document.ID,
			UserUuid:     hit.Document.UserUUID,
			Header:       hit.Document.Header,
			Body:         hit.Document.Body,
			ShortBody:    hit.Document.ShortBody,
			CreatedDate:  hit.Document.CreatedDate,
			CategoryUuid: hit.Document.CategoryUUID,
			CategoryName: hit.Document.CategoryName,
			Tags:         hit.Document.TagUUIDs,
			TagNames:     hit.Document.TagNames,
		})
	}

	return result, nil
}

func (c *Client) Upsert(note handlermodel.IndexedNote) error {
	body, err := json.Marshal(note)
	if err != nil {
		return fmt.Errorf("marshal note: %w", err)
	}

	_, err = c.doRequest(context.Background(), http.MethodPost, path.Join("collections", c.collection, "documents"), url.Values{"action": []string{"upsert"}}, body, http.StatusCreated, http.StatusOK)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) UpsertMany(notes []handlermodel.IndexedNote) error {
	var lines []string
	for _, note := range notes {
		data, err := json.Marshal(note)
		if err != nil {
			return fmt.Errorf("marshal indexed note: %w", err)
		}
		lines = append(lines, string(data))
	}

	_, err := c.doRequest(context.Background(), http.MethodPost, path.Join("collections", c.collection, "documents", "import"), url.Values{"action": []string{"upsert"}}, []byte(strings.Join(lines, "\n")), http.StatusOK)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) Delete(noteUUID, userUUID string) error {
	_, err := c.doRequest(context.Background(), http.MethodDelete, path.Join("collections", c.collection, "documents", noteUUID), nil, nil, http.StatusOK)
	if err != nil {
		var appErr *apperror.AppError
		if errors.As(err, &appErr) && errors.Is(appErr, apperror.ErrNotFound) {
			return nil
		}
		return err
	}

	return nil
}

func (c *Client) DeleteByUser(userUUID string) error {
	query := url.Values{}
	query.Set("filter_by", fmt.Sprintf("user_uuid:=`%s`", escapeFilterValue(userUUID)))

	_, err := c.doRequest(
		context.Background(),
		http.MethodDelete,
		path.Join("collections", c.collection, "documents"),
		query,
		nil,
		http.StatusOK,
	)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) ensureCollection(ctx context.Context) error {
	_, err := c.doRequest(ctx, http.MethodGet, path.Join("collections", c.collection), nil, nil, http.StatusOK)
	if err == nil {
		return nil
	}

	var appErr *apperror.AppError
	if !errors.As(err, &appErr) || !errors.Is(appErr, apperror.ErrNotFound) {
		return fmt.Errorf("check collection: %w", err)
	}

	schema := map[string]any{
		"name": c.collection,
		"fields": []map[string]any{
			{"name": "id", "type": "string"},
			{"name": "user_uuid", "type": "string", "facet": true},
			{"name": "header", "type": "string"},
			{"name": "body", "type": "string"},
			{"name": "short_body", "type": "string", "optional": true},
			{"name": "category_uuid", "type": "string", "facet": true},
			{"name": "category_name", "type": "string"},
			{"name": "tag_uuids", "type": "string[]", "facet": true, "optional": true},
			{"name": "tag_names", "type": "string[]", "optional": true},
			{"name": "created_date", "type": "int64", "sort": true},
		},
		"default_sorting_field": "created_date",
	}

	body, marshalErr := json.Marshal(schema)
	if marshalErr != nil {
		return fmt.Errorf("marshal schema: %w", marshalErr)
	}

	_, err = c.doRequest(ctx, http.MethodPost, "collections", nil, body, http.StatusCreated, http.StatusOK)
	if err != nil {
		return fmt.Errorf("create collection: %w", err)
	}

	c.logger.Info("typesense collection created")
	return nil
}

func (c *Client) doRequest(ctx context.Context, method, resource string, query url.Values, body []byte, okStatusCodes ...int) ([]byte, error) {
	reqURL := *c.baseURL
	reqURL.Path = path.Join(reqURL.Path, resource)
	if len(query) > 0 {
		reqURL.RawQuery = query.Encode()
	}

	var requestBody io.Reader
	if len(body) > 0 {
		requestBody = bytes.NewBuffer(body)
	}

	req, err := http.NewRequestWithContext(ctx, method, reqURL.String(), requestBody)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("X-TYPESENSE-API-KEY", c.apiKey)
	req.Header.Set("Accept", "application/json")
	if len(body) > 0 {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request typesense: %w", err)
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}

	for _, statusCode := range okStatusCodes {
		if resp.StatusCode == statusCode {
			return responseBody, nil
		}
	}

	if resp.StatusCode == http.StatusNotFound {
		return nil, apperror.NotFoundError("resource not found")
	}

	return nil, apperror.SystemError(fmt.Errorf("typesense returned status %d: %s", resp.StatusCode, string(responseBody)))
}

func buildFilter(userUUID, categoryUUID string, tagUUIDs []string) string {
	parts := []string{fmt.Sprintf("user_uuid:=`%s`", escapeFilterValue(userUUID))}
	if strings.TrimSpace(categoryUUID) != "" {
		parts = append(parts, fmt.Sprintf("category_uuid:=`%s`", escapeFilterValue(categoryUUID)))
	}
	if len(tagUUIDs) > 0 {
		values := make([]string, 0, len(tagUUIDs))
		for _, tagUUID := range tagUUIDs {
			if strings.TrimSpace(tagUUID) == "" {
				continue
			}
			values = append(values, fmt.Sprintf("`%s`", escapeFilterValue(tagUUID)))
		}
		if len(values) > 0 {
			parts = append(parts, fmt.Sprintf("tag_uuids:=[%s]", strings.Join(values, ",")))
		}
	}

	return strings.Join(parts, " && ")
}

func escapeFilterValue(value string) string {
	return strings.ReplaceAll(value, "`", "\\`")
}

type searchResponse struct {
	Hits []searchHit `json:"hits"`
}

type searchHit struct {
	Document handlermodel.IndexedNote `json:"document"`
}
