package events

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"

	"note_service/internal/config"
	handlermodel "note_service/internal/handlers/notes"
	"note_service/pkg/logging"
)

type Publisher interface {
	PublishNoteCreated(ctx context.Context, note handlermodel.Note) error
	PublishNoteUpdated(ctx context.Context, note handlermodel.Note) error
	PublishNoteDeleted(ctx context.Context, payload NoteDeletedPayload) error
	Close() error
}

type noopPublisher struct{}

func (noopPublisher) PublishNoteCreated(context.Context, handlermodel.Note) error  { return nil }
func (noopPublisher) PublishNoteUpdated(context.Context, handlermodel.Note) error  { return nil }
func (noopPublisher) PublishNoteDeleted(context.Context, NoteDeletedPayload) error { return nil }
func (noopPublisher) Close() error                                                 { return nil }

type rabbitPublisher struct {
	logger   logging.Logger
	url      string
	exchange string

	mu         sync.Mutex
	connection *amqp.Connection
	channel    *amqp.Channel
}

func NewPublisher(cfg *config.Config, logger logging.Logger) Publisher {
	if !cfg.RabbitMQ.Enabled {
		logger.Info("rabbitmq publisher disabled")
		return noopPublisher{}
	}

	publisher := &rabbitPublisher{
		logger:   logger.With("component", "rabbitmq_publisher"),
		url:      cfg.RabbitMQ.URL,
		exchange: cfg.RabbitMQ.Exchange,
	}

	if err := publisher.connect(); err != nil {
		logger.Warn("failed to initialize rabbitmq publisher, will retry on demand", "error", err)
	}

	return publisher
}

func (p *rabbitPublisher) PublishNoteUpdated(ctx context.Context, note handlermodel.Note) error {
	return p.publishNoteChange(ctx, NoteUpdatedEventType, note)
}

func (p *rabbitPublisher) PublishNoteCreated(ctx context.Context, note handlermodel.Note) error {
	return p.publishNoteChange(ctx, NoteCreatedEventType, note)
}

func (p *rabbitPublisher) PublishNoteDeleted(ctx context.Context, payload NoteDeletedPayload) error {
	if err := p.ensureConnection(); err != nil {
		return err
	}

	envelope := Envelope[NoteDeletedPayload]{
		EventID:      uuid.NewString(),
		EventType:    NoteDeletedEventType,
		EventVersion: 1,
		OccurredAt:   time.Now().Unix(),
		Producer:     "note_service",
		Payload:      payload,
	}

	body, err := json.Marshal(envelope)
	if err != nil {
		return fmt.Errorf("marshal note.deleted event: %w", err)
	}

	return p.publish(ctx, envelope.EventType, envelope.EventID, envelope.EventType, envelope.OccurredAt, body)
}

func (p *rabbitPublisher) publishNoteChange(ctx context.Context, eventType string, note handlermodel.Note) error {
	if err := p.ensureConnection(); err != nil {
		return err
	}

	envelope := Envelope[NoteUpdatedPayload]{
		EventID:      uuid.NewString(),
		EventType:    eventType,
		EventVersion: 1,
		OccurredAt:   time.Now().Unix(),
		Producer:     "note_service",
		Payload: NoteUpdatedPayload{
			NoteUUID:       note.Uuid,
			UserUUID:       note.UserUuid,
			WorkspaceID:    note.WorkspaceID,
			AuthorUserUUID: note.AuthorUserUUID,
			CategoryUUID:   note.CategoryUuid,
			Header:         note.Header,
			Body:           note.Body,
			ShortBody:      note.ShortBody,
			Tags:           note.Tags,
			Event:          note.Event,
			CreatedAt:      note.CreatedDate,
			UpdatedAt:      note.UpdatedAt,
		},
	}

	body, err := json.Marshal(envelope)
	if err != nil {
		return fmt.Errorf("marshal %s event: %w", eventType, err)
	}

	return p.publish(ctx, envelope.EventType, envelope.EventID, envelope.EventType, envelope.OccurredAt, body)
}

func (p *rabbitPublisher) publish(ctx context.Context, routingKey, eventID, eventType string, occurredAt int64, body []byte) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.channel == nil {
		return fmt.Errorf("rabbitmq channel is not initialized")
	}

	if publishErr := p.channel.PublishWithContext(
		ctx,
		p.exchange,
		routingKey,
		false,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			MessageId:    eventID,
			Timestamp:    time.Unix(occurredAt, 0),
			Type:         eventType,
			Body:         body,
		},
	); publishErr != nil {
		return fmt.Errorf("publish %s event: %w", eventType, publishErr)
	}

	return nil
}

func (p *rabbitPublisher) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	var closeErr error
	if p.channel != nil {
		if err := p.channel.Close(); err != nil {
			closeErr = err
		}
		p.channel = nil
	}
	if p.connection != nil {
		if err := p.connection.Close(); err != nil && closeErr == nil {
			closeErr = err
		}
		p.connection = nil
	}

	return closeErr
}

func (p *rabbitPublisher) ensureConnection() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.connection != nil && !p.connection.IsClosed() && p.channel != nil {
		return nil
	}

	return p.connectLocked()
}

func (p *rabbitPublisher) connect() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	return p.connectLocked()
}

func (p *rabbitPublisher) connectLocked() error {
	if p.channel != nil {
		_ = p.channel.Close()
		p.channel = nil
	}
	if p.connection != nil {
		_ = p.connection.Close()
		p.connection = nil
	}

	connection, err := amqp.Dial(p.url)
	if err != nil {
		return fmt.Errorf("dial rabbitmq: %w", err)
	}

	channel, err := connection.Channel()
	if err != nil {
		_ = connection.Close()
		return fmt.Errorf("open rabbitmq channel: %w", err)
	}

	if err = channel.ExchangeDeclare(
		p.exchange,
		"topic",
		true,
		false,
		false,
		false,
		nil,
	); err != nil {
		_ = channel.Close()
		_ = connection.Close()
		return fmt.Errorf("declare exchange %s: %w", p.exchange, err)
	}

	p.connection = connection
	p.channel = channel
	p.logger.Info("rabbitmq publisher connected", "exchange", p.exchange)
	return nil
}
