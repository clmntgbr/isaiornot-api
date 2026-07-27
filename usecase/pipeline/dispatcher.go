package pipeline

import (
	"context"
	"fmt"

	"go-api/domain/repository"
	"go-api/infrastructure/config"
	"go-api/infrastructure/messaging/rabbitmq"
)

type Dispatcher struct {
	config          *config.Config
	mediaRepo       *repository.MediaRepository
	publisher       rabbitmq.Publisher
	finalizeUseCase *AggregateScanUseCase
}

func NewDispatcher(
	config *config.Config,
	mediaRepo *repository.MediaRepository,
	publisher rabbitmq.Publisher,
	finalizeUseCase *AggregateScanUseCase,
) *Dispatcher {
	return &Dispatcher{
		config:          config,
		mediaRepo:       mediaRepo,
		publisher:       publisher,
		finalizeUseCase: finalizeUseCase,
	}
}

func (d *Dispatcher) HandleAnalyzeRequest(ctx context.Context, message rabbitmq.AnalyzeMessage) error {
	return d.publisher.Publish(ctx, d.config.MetadataAnalyzeQueueName, message)
}

func (d *Dispatcher) HandleStageDone(ctx context.Context, message rabbitmq.StageDoneMessage) error {
	media, err := (*d.mediaRepo).GetByID(ctx, message.MediaID)
	if err != nil {
		return fmt.Errorf("media not found: %w", err)
	}

	next := d.nextStage(message.Stage)
	if next == "" {
		return d.finalizeUseCase.Execute(ctx, media.ID)
	}

	analyzeMessage := rabbitmq.AnalyzeMessage{
		UserID:       media.UserID,
		MediaID:      media.ID,
		MediaKey:     media.Key,
		ThumbnailKey: media.Thumbnail,
	}

	queueName := d.stageAnalyzeQueue(next)
	if queueName == "" {
		return fmt.Errorf("unknown stage %q", next)
	}

	return d.publisher.Publish(ctx, queueName, analyzeMessage)
}

func (d *Dispatcher) nextStage(current string) string {
	order := []string{"metadata", "heuristics", "ai_model"}
	for i, stage := range order {
		if stage == current && i+1 < len(order) {
			return order[i+1]
		}
	}

	return ""
}

func (d *Dispatcher) stageAnalyzeQueue(stage string) string {
	switch stage {
	case "metadata":
		return d.config.MetadataAnalyzeQueueName
	case "heuristics":
		return d.config.HeuristicsAnalyzeQueueName
	case "ai_model":
		return d.config.AiModelAnalyzeQueueName
	default:
		return ""
	}
}

func (d *Dispatcher) StageFailedQueue(stage string) string {
	switch stage {
	case "metadata":
		return d.config.MetadataFailedQueueName
	case "heuristics":
		return d.config.HeuristicsFailedQueueName
	case "ai_model":
		return d.config.AiModelFailedQueueName
	default:
		return ""
	}
}

func (d *Dispatcher) StageDoneQueue(stage string) string {
	switch stage {
	case "metadata":
		return d.config.MetadataDoneQueueName
	case "heuristics":
		return d.config.HeuristicsDoneQueueName
	case "ai_model":
		return d.config.AiModelDoneQueueName
	default:
		return ""
	}
}

func (d *Dispatcher) PublishStageDone(ctx context.Context, queueName string, message rabbitmq.StageDoneMessage) error {
	return d.publisher.Publish(ctx, queueName, message)
}

func (d *Dispatcher) PublishFailed(ctx context.Context, queueName string, message rabbitmq.FailedMessage) error {
	return d.publisher.Publish(ctx, queueName, message)
}
