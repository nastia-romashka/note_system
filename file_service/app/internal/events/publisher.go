package events

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"

	"file_service/internal/config"
	domainfile "file_service/internal/file"
	"file_service/pkg/logging"
)

type Publisher interface {
	PublishFileUploaded(ctx context.Context, file domainfile.FileInfo) error
	PublishFileDeleted(ctx context.Context, payload FileDeletedPayload) error
	Close() error
}

type noopPublisher struct{}

func (noopPublisher) PublishFileUploaded(context.Context, domainfile.FileInfo) error { return nil }
func (noopPublisher) PublishFileDeleted(context.Context, FileDeletedPayload) error   { return nil }
func (noopPublisher) Close() error                                                   { return nil }

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

func (p *rabbitPublisher) PublishFileUploaded(ctx context.Context, file domainfile.FileInfo) error {
	envelope := Envelope[FileUploadedPayload]{
		EventID:      uuid.NewString(),
		EventType:    FileUploadedEventType,
		EventVersion: 1,
		OccurredAt:   time.Now().Unix(),
		Producer:     "file_service",
		Payload:      NewFileUploadedPayload(file),
	}
	return p.publishEnvelope(ctx, envelope)
}

func (p *rabbitPublisher) PublishFileDeleted(ctx context.Context, payload FileDeletedPayload) error {
	envelope := Envelope[FileDeletedPayload]{
		EventID:      uuid.NewString(),
		EventType:    FileDeletedEventType,
		EventVersion: 1,
		OccurredAt:   time.Now().Unix(),
		Producer:     "file_service",
		Payload:      payload,
	}
	return p.publishEnvelope(ctx, envelope)
}

func (p *rabbitPublisher) publishEnvelope(ctx context.Context, envelope any) error {
	if err := p.ensureConnection(); err != nil {
		return err
	}

	body, err := json.Marshal(envelope)
	if err != nil {
		return fmt.Errorf("marshal event: %w", err)
	}

	rawEnvelope, ok := envelope.(interface {
		GetEventID() string
	})
	if ok {
		_ = rawEnvelope
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	if p.channel == nil {
		return fmt.Errorf("rabbitmq channel is not initialized")
	}

	var eventID, eventType string
	var occurredAt int64
	switch event := envelope.(type) {
	case Envelope[FileUploadedPayload]:
		eventID, eventType, occurredAt = event.EventID, event.EventType, event.OccurredAt
	case Envelope[FileDeletedPayload]:
		eventID, eventType, occurredAt = event.EventID, event.EventType, event.OccurredAt
	default:
		return fmt.Errorf("unsupported event envelope type")
	}

	if err = p.channel.PublishWithContext(
		ctx,
		p.exchange,
		eventType,
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
	); err != nil {
		return fmt.Errorf("publish %s event: %w", eventType, err)
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
	if err = channel.ExchangeDeclare(p.exchange, "topic", true, false, false, false, nil); err != nil {
		_ = channel.Close()
		_ = connection.Close()
		return fmt.Errorf("declare exchange %s: %w", p.exchange, err)
	}
	p.connection = connection
	p.channel = channel
	p.logger.Info("rabbitmq publisher connected", "exchange", p.exchange)
	return nil
}
