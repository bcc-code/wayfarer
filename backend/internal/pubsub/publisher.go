package pubsub

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"cloud.google.com/go/pubsub"
)

// Publisher handles publishing messages to GCP Pub/Sub
type Publisher struct {
	client  *pubsub.Client
	topic   *pubsub.Topic
	enabled bool
	logger  *slog.Logger
}

// PublisherConfig holds configuration for the Pub/Sub publisher
type PublisherConfig struct {
	Enabled   bool
	ProjectID string
	TopicID   string
}

// NewPublisher creates a new Pub/Sub publisher
func NewPublisher(ctx context.Context, cfg PublisherConfig, logger *slog.Logger) (*Publisher, error) {
	if !cfg.Enabled {
		logger.Info("Pub/Sub publisher disabled")
		return &Publisher{
			enabled: false,
			logger:  logger,
		}, nil
	}

	client, err := pubsub.NewClient(ctx, cfg.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("failed to create pub/sub client: %w", err)
	}

	topic := client.Topic(cfg.TopicID)
	exists, err := topic.Exists(ctx)
	if err != nil {
		client.Close()
		return nil, fmt.Errorf("failed to check topic existence: %w", err)
	}
	if !exists {
		client.Close()
		return nil, fmt.Errorf("topic %s does not exist", cfg.TopicID)
	}

	logger.Info("Pub/Sub publisher initialized",
		"project_id", cfg.ProjectID,
		"topic_id", cfg.TopicID,
	)

	return &Publisher{
		client:  client,
		topic:   topic,
		enabled: true,
		logger:  logger,
	}, nil
}

// PublishBulkOperation publishes a bulk operation message to Pub/Sub
func (p *Publisher) PublishBulkOperation(ctx context.Context, msg BulkOperationMessage) (string, error) {
	if !p.enabled {
		p.logger.Warn("Pub/Sub disabled, cannot publish message",
			"job_id", msg.JobID,
			"operation_type", msg.OperationType,
		)
		return "", fmt.Errorf("pub/sub is disabled")
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return "", fmt.Errorf("failed to marshal message: %w", err)
	}

	result := p.topic.Publish(ctx, &pubsub.Message{
		Data: data,
		Attributes: map[string]string{
			"job_id":         msg.JobID,
			"operation_type": string(msg.OperationType),
		},
	})

	messageID, err := result.Get(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to publish message: %w", err)
	}

	p.logger.Info("Published bulk operation message",
		"job_id", msg.JobID,
		"operation_type", msg.OperationType,
		"message_id", messageID,
	)

	return messageID, nil
}

// IsEnabled returns whether the publisher is enabled
func (p *Publisher) IsEnabled() bool {
	return p.enabled
}

// Close closes the Pub/Sub client
func (p *Publisher) Close() error {
	if p.client != nil {
		return p.client.Close()
	}
	return nil
}
