package events

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"

	"user_service/internal/config"
	"user_service/pkg/logging"
)

type Publisher interface {
	PublishWorkspaceInviteAccepted(ctx context.Context, payload WorkspaceInviteAcceptedPayload) error
	Close() error
}

type noopPublisher struct{}

func (noopPublisher) PublishWorkspaceInviteAccepted(context.Context, WorkspaceInviteAcceptedPayload) error {
	return nil
}
func (noopPublisher) Close() error { return nil }

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

func (p *rabbitPublisher) PublishWorkspaceInviteAccepted(ctx context.Context, payload WorkspaceInviteAcceptedPayload) error {
	if err := p.ensureConnection(); err != nil {
		return err
	}

	envelope := Envelope[WorkspaceInviteAcceptedPayload]{
		EventID:      uuid.NewString(),
		EventType:    WorkspaceInviteAcceptedEventType,
		EventVersion: 1,
		OccurredAt:   time.Now().Unix(),
		Producer:     "user_service",
		Payload:      payload,
	}
	body, err := json.Marshal(envelope)
	if err != nil {
		return fmt.Errorf("marshal workspace invite accepted event: %w", err)
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	if p.channel == nil {
		return fmt.Errorf("rabbitmq channel is not initialized")
	}

	if err = p.channel.PublishWithContext(
		ctx,
		p.exchange,
		envelope.EventType,
		false,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			MessageId:    envelope.EventID,
			Timestamp:    time.Unix(envelope.OccurredAt, 0),
			Type:         envelope.EventType,
			Body:         body,
		},
	); err != nil {
		return fmt.Errorf("publish workspace invite accepted event: %w", err)
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
