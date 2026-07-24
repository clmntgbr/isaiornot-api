package rabbitmq

import (
	"fmt"
	"log"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	dialMaxAttempts = 30
	dialRetryDelay  = 2 * time.Second
)

func dialWithRetry(url string) (*amqp.Connection, error) {
	var lastErr error

	for attempt := 1; attempt <= dialMaxAttempts; attempt++ {
		conn, err := amqp.Dial(url)
		if err == nil {
			return conn, nil
		}

		lastErr = err
		log.Printf(
			"failed to connect to RabbitMQ (attempt %d/%d): %v",
			attempt,
			dialMaxAttempts,
			err,
		)

		if attempt < dialMaxAttempts {
			time.Sleep(dialRetryDelay)
		}
	}

	return nil, fmt.Errorf(
		"failed to connect to RabbitMQ after %d attempts: %w",
		dialMaxAttempts,
		lastErr,
	)
}
