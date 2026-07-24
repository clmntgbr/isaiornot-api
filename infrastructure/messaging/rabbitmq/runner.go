package rabbitmq

import (
	"context"
	"fmt"
	"log"

	"go-api/infrastructure/config"

	amqp "github.com/rabbitmq/amqp091-go"
)

type MessageHandler interface {
	HandleMessage(ctx context.Context, delivery *amqp.Delivery) error
}

type Worker struct {
	env       *config.Config
	queueName string
	handler   MessageHandler

	conn    *amqp.Connection
	channel *amqp.Channel
}

func NewWorker(
	env *config.Config,
	queueName string,
	handler MessageHandler,
) *Worker {
	return &Worker{
		env:       env,
		queueName: queueName,
		handler:   handler,
	}
}

func (w *Worker) Start() error {
	conn, err := dialWithRetry(w.env.RabbitMQURL)
	if err != nil {
		return fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	w.conn = conn

	channel, err := conn.Channel()
	if err != nil {
		return fmt.Errorf("failed to open RabbitMQ channel: %w", err)
	}

	w.channel = channel

	if err := w.channel.ExchangeDeclare(
		w.env.ExchangeName,
		"topic",
		true,
		false,
		false,
		false,
		nil,
	); err != nil {
		return fmt.Errorf(
			"failed to declare exchange %q: %w",
			w.env.ExchangeName,
			err,
		)
	}

	queue, err := w.channel.QueueDeclare(
		w.queueName,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf(
			"failed to declare queue %q: %w",
			w.queueName,
			err,
		)
	}

	messages, err := w.channel.Consume(
		queue.Name,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf(
			"failed to consume queue %q: %w",
			queue.Name,
			err,
		)
	}

	for message := range messages {
		if err := w.handler.HandleMessage(context.Background(), &message); err != nil {
			log.Printf(
				"rejected message (routing key: %q): %v, body: %s",
				message.RoutingKey,
				err,
				message.Body,
			)

			if nackErr := message.Nack(false, false); nackErr != nil {
				log.Printf("failed to nack message: %v", nackErr)
			}

			continue
		}

		if ackErr := message.Ack(false); ackErr != nil {
			log.Printf("failed to ack message: %v", ackErr)
		}
	}

	return nil
}

func (w *Worker) Stop() error {
	if w.channel != nil {
		if err := w.channel.Close(); err != nil {
			return err
		}
	}

	if w.conn != nil {
		if err := w.conn.Close(); err != nil {
			return err
		}
	}

	return nil
}
