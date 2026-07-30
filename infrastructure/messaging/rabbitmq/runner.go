package rabbitmq

import (
	"context"
	"fmt"
	"log"
	"sync"

	"go-api/infrastructure/config"

	amqp "github.com/rabbitmq/amqp091-go"
)

type MessageHandler interface {
	HandleMessage(ctx context.Context, delivery *amqp.Delivery) error
}

type Worker struct {
	env         *config.Config
	queueName   string
	handler     MessageHandler
	concurrency int

	conn    *amqp.Connection
	channel *amqp.Channel
	sem     chan struct{}
	ackMu   sync.Mutex
	wg      sync.WaitGroup
}

func NewWorker(
	env *config.Config,
	queueName string,
	handler MessageHandler,
	concurrency int,
) *Worker {
	if concurrency < 1 {
		concurrency = 1
	}

	return &Worker{
		env:         env,
		queueName:   queueName,
		handler:     handler,
		concurrency: concurrency,
		sem:         make(chan struct{}, concurrency),
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

	if err := w.channel.Qos(w.concurrency, 0, false); err != nil {
		return fmt.Errorf("failed to set QoS for queue %q: %w", w.queueName, err)
	}

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
		msg := message
		w.wg.Add(1)
		w.sem <- struct{}{}
		go func() {
			defer func() { <-w.sem }()
			w.process(&msg)
		}()
	}

	w.wg.Wait()
	return nil
}

func (w *Worker) process(message *amqp.Delivery) {
	defer w.wg.Done()

	if err := w.handler.HandleMessage(context.Background(), message); err != nil {
		log.Printf(
			"rejected message (routing key: %q): %v, body: %s",
			message.RoutingKey,
			err,
			message.Body,
		)

		w.ackMu.Lock()
		defer w.ackMu.Unlock()
		if nackErr := message.Nack(false, false); nackErr != nil {
			log.Printf("failed to nack message: %v", nackErr)
		}
		return
	}

	w.ackMu.Lock()
	defer w.ackMu.Unlock()
	if ackErr := message.Ack(false); ackErr != nil {
		log.Printf("failed to ack message: %v", ackErr)
	}
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
