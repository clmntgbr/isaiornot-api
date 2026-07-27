package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"

	"go-api/domain/port"
	"go-api/infrastructure/config"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Publisher = port.MessagePublisher

type publisher struct {
	env     *config.Config
	channel *amqp.Channel
}

func NewPublisher(env *config.Config, channel *amqp.Channel) Publisher {
	return &publisher{
		env:     env,
		channel: channel,
	}
}

func NewPublisherFromEnv(env *config.Config) (Publisher, error) {
	conn, err := dialWithRetry(env.RabbitMQURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ at %s: %w", env.RabbitMQURL, err)
	}

	ch, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("failed to open RabbitMQ channel: %w", err)
	}

	return NewPublisher(env, ch), nil
}

func (p *publisher) Publish(ctx context.Context, queueName string, message any) error {
	payload := MessagePayload{
		SecretKey: p.env.RabbitMQSecretKey,
		Message:   message,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	return p.channel.PublishWithContext(
		ctx,
		"",
		queueName,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)
}
