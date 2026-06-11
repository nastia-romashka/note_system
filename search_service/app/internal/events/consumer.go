package events

import (
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

	amqp "github.com/rabbitmq/amqp091-go"

	"search_service/internal/config"
	handlermodel "search_service/internal/handlers/notes"
	"search_service/pkg/logging"
)

const searchQueueName = "search.domain-events"

var errNotFound = errors.New("resource not found")

type NotesIndexer interface {
	Upsert(note handlermodel.IndexedNote) error
	UpsertMany(notes []handlermodel.IndexedNote) error
	Delete(noteUUID string) error
}

type Consumer struct {
	logger      logging.Logger
	config      *config.Config
	repository  NotesIndexer
	httpClient  *http.Client
	connection  *amqp.Connection
	channel     *amqp.Channel
}

type Envelope[T any] struct {
	EventID      string `json:"event_id"`
	EventType    string `json:"event_type"`
	EventVersion int    `json:"event_version"`
	OccurredAt   int64  `json:"occurred_at"`
	Producer     string `json:"producer"`
	Payload      T      `json:"payload"`
}

type noteEventPayload struct {
	NoteUUID    string   `json:"note_uuid"`
	UserUUID    string   `json:"user_uuid"`
	WorkspaceID string   `json:"workspace_id"`
	CategoryUUID string  `json:"category_uuid,omitempty"`
	Header      string   `json:"header,omitempty"`
	Body        string   `json:"body,omitempty"`
	Tags        []string `json:"tags,omitempty"`
}

type fileEventPayload struct {
	FileID      string `json:"file_id"`
	UserUUID    string `json:"user_uuid"`
	WorkspaceID string `json:"workspace_id"`
	NoteUUID    string `json:"note_uuid"`
	Name        string `json:"name,omitempty"`
}

type categoryUpdatedPayload struct {
	CategoryUUID   string `json:"category_uuid"`
	WorkspaceID    string `json:"workspace_id"`
	ActorUserUUID  string `json:"actor_user_uuid"`
	Name           string `json:"name"`
}

type noteDTO struct {
	Uuid         string   `json:"uuid"`
	WorkspaceID  string   `json:"workspace_id"`
	Header       string   `json:"header"`
	Body         string   `json:"body"`
	ShortBody    string   `json:"short_body,omitempty"`
	CreatedDate  int64    `json:"created_date"`
	UpdatedAt    int64    `json:"updated_at"`
	CategoryUuid string   `json:"category_uuid"`
	Tags         []string `json:"tags,omitempty"`
}

type tagDTO struct {
	Uuid string `json:"uuid"`
	Name string `json:"name"`
}

type categoryDTO struct {
	Uuid     string        `json:"uuid"`
	Name     string        `json:"name"`
	Children []categoryDTO `json:"children,omitempty"`
}

type fileDTO struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	WorkspaceID string `json:"workspace_id,omitempty"`
	NoteUUID    string `json:"note_uuid"`
	ContentType string `json:"content_type"`
	Size        int64  `json:"size"`
}

func NewConsumer(cfg *config.Config, logger logging.Logger, repository NotesIndexer) *Consumer {
	return &Consumer{
		logger:     logger.With("component", "domain_events_consumer"),
		config:     cfg,
		repository: repository,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

func (c *Consumer) Start(ctx context.Context) error {
	if !c.config.RabbitMQ.Enabled {
		c.logger.Info("rabbitmq consumer disabled")
		return nil
	}

	conn, err := amqp.Dial(c.config.RabbitMQ.URL)
	if err != nil {
		return fmt.Errorf("dial rabbitmq: %w", err)
	}
	ch, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		return fmt.Errorf("open rabbitmq channel: %w", err)
	}
	if err = ch.ExchangeDeclare(c.config.RabbitMQ.Exchange, "topic", true, false, false, false, nil); err != nil {
		_ = ch.Close()
		_ = conn.Close()
		return fmt.Errorf("declare exchange: %w", err)
	}
	if _, err = ch.QueueDeclare(searchQueueName, true, false, false, false, nil); err != nil {
		_ = ch.Close()
		_ = conn.Close()
		return fmt.Errorf("declare queue: %w", err)
	}

	for _, routingKey := range []string{
		"note.created",
		"note.updated",
		"note.deleted",
		"file.uploaded",
		"file.deleted",
		"category.updated",
	} {
		if err = ch.QueueBind(searchQueueName, routingKey, c.config.RabbitMQ.Exchange, false, nil); err != nil {
			_ = ch.Close()
			_ = conn.Close()
			return fmt.Errorf("bind queue for %s: %w", routingKey, err)
		}
	}

	if err = ch.Qos(10, 0, false); err != nil {
		_ = ch.Close()
		_ = conn.Close()
		return fmt.Errorf("configure qos: %w", err)
	}

	deliveries, err := ch.Consume(searchQueueName, "", false, false, false, false, nil)
	if err != nil {
		_ = ch.Close()
		_ = conn.Close()
		return fmt.Errorf("consume queue: %w", err)
	}

	c.connection = conn
	c.channel = ch
	c.logger.Info("search domain events consumer connected", "queue", searchQueueName)

	go func() {
		<-ctx.Done()
		_ = c.Close()
	}()

	go func() {
		for delivery := range deliveries {
			if err := c.handleDelivery(ctx, delivery); err != nil {
				c.logger.Warn("failed to handle domain event", "error", err)
			}
		}
	}()

	return nil
}

func (c *Consumer) Close() error {
	var closeErr error
	if c.channel != nil {
		if err := c.channel.Close(); err != nil {
			closeErr = err
		}
		c.channel = nil
	}
	if c.connection != nil {
		if err := c.connection.Close(); err != nil && closeErr == nil {
			closeErr = err
		}
		c.connection = nil
	}
	return closeErr
}

func (c *Consumer) handleDelivery(ctx context.Context, delivery amqp.Delivery) error {
	var envelope Envelope[json.RawMessage]
	if err := json.Unmarshal(delivery.Body, &envelope); err != nil {
		_ = delivery.Ack(false)
		return fmt.Errorf("decode envelope: %w", err)
	}

	var err error
	switch envelope.EventType {
	case "note.created", "note.updated":
		var payload noteEventPayload
		if err = json.Unmarshal(envelope.Payload, &payload); err == nil {
			err = c.reindexSingleNote(ctx, payload.NoteUUID, payload.UserUUID, payload.WorkspaceID)
		}
	case "note.deleted":
		var payload noteEventPayload
		if err = json.Unmarshal(envelope.Payload, &payload); err == nil {
			err = c.repository.Delete(payload.NoteUUID)
		}
	case "file.uploaded", "file.deleted":
		var payload fileEventPayload
		if err = json.Unmarshal(envelope.Payload, &payload); err == nil {
			err = c.reindexSingleNote(ctx, payload.NoteUUID, payload.UserUUID, payload.WorkspaceID)
		}
	case "category.updated":
		var payload categoryUpdatedPayload
		if err = json.Unmarshal(envelope.Payload, &payload); err == nil {
			err = c.reindexCategory(ctx, payload)
		}
	}

	if err == nil {
		return delivery.Ack(false)
	}
	if errors.Is(err, errNotFound) {
		_ = delivery.Ack(false)
		return nil
	}
	_ = delivery.Ack(false)
	return err
}

func (c *Consumer) reindexSingleNote(ctx context.Context, noteUUID, userUUID, workspaceID string) error {
	note, err := c.fetchNote(ctx, noteUUID, userUUID, workspaceID)
	if err != nil {
		return err
	}
	categories, err := c.fetchCategories(ctx, workspaceID)
	if err != nil {
		return err
	}
	tags, err := c.fetchTags(ctx, note.Tags, userUUID, workspaceID)
	if err != nil {
		return err
	}
	files, err := c.fetchFiles(ctx, note.Uuid, userUUID, workspaceID)
	if err != nil {
		return err
	}
	document, err := buildIndexedNote(note, categories, tags, files)
	if err != nil {
		return err
	}
	return c.repository.Upsert(document)
}

func (c *Consumer) reindexCategory(ctx context.Context, payload categoryUpdatedPayload) error {
	if strings.TrimSpace(payload.ActorUserUUID) == "" {
		return nil
	}
	notes, err := c.fetchNotesByCategory(ctx, payload.CategoryUUID, payload.ActorUserUUID, payload.WorkspaceID)
	if err != nil {
		return err
	}
	if len(notes) == 0 {
		return nil
	}
	categories, err := c.fetchCategories(ctx, payload.WorkspaceID)
	if err != nil {
		return err
	}
	tags, err := c.fetchTags(ctx, collectTagUUIDs(notes), payload.ActorUserUUID, payload.WorkspaceID)
	if err != nil {
		return err
	}
	filesByNote := make(map[string][]fileDTO, len(notes))
	for _, note := range notes {
		files, filesErr := c.fetchFiles(ctx, note.Uuid, payload.ActorUserUUID, payload.WorkspaceID)
		if filesErr != nil {
			return filesErr
		}
		filesByNote[note.Uuid] = files
	}
	documents, err := buildIndexedNotes(notes, categories, tags, filesByNote)
	if err != nil {
		return err
	}
	return c.repository.UpsertMany(documents)
}

func (c *Consumer) fetchNote(ctx context.Context, noteUUID, userUUID, workspaceID string) (noteDTO, error) {
	var note noteDTO
	err := c.getJSON(ctx, c.config.NoteService.URL, path.Join("notes", noteUUID), url.Values{
		"user_uuid":    []string{userUUID},
		"workspace_id": []string{workspaceID},
	}, &note)
	return note, err
}

func (c *Consumer) fetchNotesByCategory(ctx context.Context, categoryUUID, userUUID, workspaceID string) ([]noteDTO, error) {
	var notes []noteDTO
	err := c.getJSON(ctx, c.config.NoteService.URL, "notes", url.Values{
		"category_uuid": []string{categoryUUID},
		"user_uuid":     []string{userUUID},
		"workspace_id":  []string{workspaceID},
	}, &notes)
	return notes, err
}

func (c *Consumer) fetchTags(ctx context.Context, tagUUIDs []string, userUUID, workspaceID string) ([]tagDTO, error) {
	if len(tagUUIDs) == 0 {
		return nil, nil
	}
	query := url.Values{
		"user_uuid":    []string{userUUID},
		"workspace_id": []string{workspaceID},
	}
	for _, tagUUID := range tagUUIDs {
		if strings.TrimSpace(tagUUID) == "" {
			continue
		}
		query.Add("id", tagUUID)
	}
	var tags []tagDTO
	err := c.getJSON(ctx, c.config.NoteService.URL, "tags", query, &tags)
	return tags, err
}

func (c *Consumer) fetchCategories(ctx context.Context, workspaceID string) ([]categoryDTO, error) {
	var categories []categoryDTO
	err := c.getJSON(ctx, c.config.CategoryService.URL, "categories", url.Values{
		"workspace_id": []string{workspaceID},
	}, &categories)
	return categories, err
}

func (c *Consumer) fetchFiles(ctx context.Context, noteUUID, userUUID, workspaceID string) ([]fileDTO, error) {
	var files []fileDTO
	err := c.getJSON(ctx, c.config.FileService.URL, path.Join("api", "files"), url.Values{
		"note_uuid":    []string{noteUUID},
		"user_uuid":    []string{userUUID},
		"workspace_id": []string{workspaceID},
	}, &files)
	return files, err
}

func (c *Consumer) getJSON(ctx context.Context, baseURL, resource string, query url.Values, dst any) error {
	base, err := url.Parse(strings.TrimSpace(baseURL))
	if err != nil {
		return fmt.Errorf("parse service url: %w", err)
	}
	base.Path = path.Join(base.Path, resource)
	base.RawQuery = query.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, base.String(), nil)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("perform request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return errNotFound
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status %d from %s: %s", resp.StatusCode, base.String(), string(body))
	}

	if err = json.NewDecoder(resp.Body).Decode(dst); err != nil {
		return fmt.Errorf("decode response from %s: %w", base.String(), err)
	}
	return nil
}

func buildIndexedNote(note noteDTO, categories []categoryDTO, tags []tagDTO, files []fileDTO) (handlermodel.IndexedNote, error) {
	categoryName, ok := findCategoryName(categories, note.CategoryUuid)
	if !ok {
		return handlermodel.IndexedNote{}, fmt.Errorf("category %s not found in workspace category tree", note.CategoryUuid)
	}

	tagNamesByID := make(map[string]string, len(tags))
	for _, tag := range tags {
		tagNamesByID[tag.Uuid] = tag.Name
	}

	tagNames := make([]string, 0, len(note.Tags))
	for _, tagUUID := range note.Tags {
		if name, ok := tagNamesByID[tagUUID]; ok {
			tagNames = append(tagNames, name)
		}
	}

	return handlermodel.IndexedNote{
		ID:            note.Uuid,
		WorkspaceID:   note.WorkspaceID,
		Header:        note.Header,
		Body:          note.Body,
		ShortBody:     buildShortBody(note),
		CategoryUUID:  note.CategoryUuid,
		CategoryName:  categoryName,
		TagUUIDs:      note.Tags,
		TagNames:      tagNames,
		FileNamesText: buildFileNamesText(files),
		CreatedAt:     note.CreatedDate,
		UpdatedAt:     note.UpdatedAt,
	}, nil
}

func buildIndexedNotes(notes []noteDTO, categories []categoryDTO, tags []tagDTO, filesByNote map[string][]fileDTO) ([]handlermodel.IndexedNote, error) {
	result := make([]handlermodel.IndexedNote, 0, len(notes))
	for _, note := range notes {
		document, err := buildIndexedNote(note, categories, tags, filesByNote[note.Uuid])
		if err != nil {
			return nil, err
		}
		result = append(result, document)
	}
	return result, nil
}

func buildShortBody(note noteDTO) string {
	if strings.TrimSpace(note.ShortBody) != "" {
		return note.ShortBody
	}
	body := strings.Join(strings.Fields(strings.TrimSpace(note.Body)), " ")
	if len(body) <= 120 {
		return body
	}
	if len(body) == 0 {
		return ""
	}
	return body[:117] + "..."
}

func findCategoryName(categories []categoryDTO, categoryUUID string) (string, bool) {
	for _, category := range categories {
		if category.Uuid == categoryUUID {
			return category.Name, true
		}
		if name, ok := findCategoryName(category.Children, categoryUUID); ok {
			return name, true
		}
	}
	return "", false
}

func collectTagUUIDs(notes []noteDTO) []string {
	seen := make(map[string]struct{})
	result := make([]string, 0)
	for _, note := range notes {
		for _, tagUUID := range note.Tags {
			if strings.TrimSpace(tagUUID) == "" {
				continue
			}
			if _, ok := seen[tagUUID]; ok {
				continue
			}
			seen[tagUUID] = struct{}{}
			result = append(result, tagUUID)
		}
	}
	return result
}

func buildFileNamesText(files []fileDTO) string {
	if len(files) == 0 {
		return ""
	}
	seen := make(map[string]struct{}, len(files))
	names := make([]string, 0, len(files))
	for _, file := range files {
		name := strings.TrimSpace(file.Name)
		if name == "" {
			continue
		}
		key := strings.ToLower(name)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		names = append(names, name)
	}
	return strings.Join(names, " ")
}
