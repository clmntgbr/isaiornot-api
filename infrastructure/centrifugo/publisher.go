package centrifugo

import (
	"context"
	"fmt"
	"strings"

	"go-api/domain/realtime"
	"go-api/infrastructure/config"

	"github.com/centrifugal/gocent/v3"
	"github.com/google/uuid"
)

type Publisher struct {
	client *gocent.Client
}

func NewPublisher(env *config.Config) *Publisher {
	return &Publisher{
		client: gocent.New(gocent.Config{
			Addr: apiEndpoint(env.CentrifugoURL),
			Key:  env.CentrifugoAPIKey,
		}),
	}
}

func apiEndpoint(baseURL string) string {
	baseURL = strings.TrimRight(baseURL, "/")
	if strings.HasSuffix(baseURL, "/api") {
		return baseURL
	}

	return baseURL + "/api"
}

func (p *Publisher) PublishToUser(ctx context.Context, userID uuid.UUID, event realtime.MediaEvent) error {
	payload, err := event.Marshal()
	if err != nil {
		return fmt.Errorf("failed to marshal centrifugo event: %w", err)
	}

	channel := UserChannel(userID)
	if _, err := p.client.Publish(ctx, channel, payload); err != nil {
		return fmt.Errorf("failed to publish to centrifugo channel %q: %w", channel, err)
	}

	return nil
}

func (p *Publisher) PublishSubscriptionToUser(ctx context.Context, userID uuid.UUID, event realtime.SubscriptionEvent) error {
	payload, err := event.Marshal()
	if err != nil {
		return fmt.Errorf("failed to marshal centrifugo event: %w", err)
	}

	channel := UserChannel(userID)
	if _, err := p.client.Publish(ctx, channel, payload); err != nil {
		return fmt.Errorf("failed to publish to centrifugo channel %q: %w", channel, err)
	}

	return nil
}

func (p *Publisher) PublishUserToUser(ctx context.Context, userID uuid.UUID, event realtime.UserEvent) error {
	payload, err := event.Marshal()
	if err != nil {
		return fmt.Errorf("failed to marshal centrifugo event: %w", err)
	}

	channel := UserChannel(userID)
	if _, err := p.client.Publish(ctx, channel, payload); err != nil {
		return fmt.Errorf("failed to publish to centrifugo channel %q: %w", channel, err)
	}

	return nil
}

func (p *Publisher) PublishQuotaToUser(ctx context.Context, userID uuid.UUID, event realtime.QuotaEvent) error {
	payload, err := event.Marshal()
	if err != nil {
		return fmt.Errorf("failed to marshal centrifugo event: %w", err)
	}

	channel := UserChannel(userID)
	if _, err := p.client.Publish(ctx, channel, payload); err != nil {
		return fmt.Errorf("failed to publish to centrifugo channel %q: %w", channel, err)
	}

	return nil
}
