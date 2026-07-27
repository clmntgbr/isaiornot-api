package pipeline

import (
	"context"
	"fmt"

	"go-api/domain/messaging"
	"go-api/domain/port"
	"go-api/domain/repository"
)

type Dispatcher struct {
	queues          port.AnalyzeQueues
	mediaRepo       repository.MediaRepository
	publisher       port.MessagePublisher
	finalizeUseCase *AggregateScanUseCase
}

func NewDispatcher(
	queues port.AnalyzeQueues,
	mediaRepo repository.MediaRepository,
	publisher port.MessagePublisher,
	finalizeUseCase *AggregateScanUseCase,
) *Dispatcher {
	return &Dispatcher{
		queues:          queues,
		mediaRepo:       mediaRepo,
		publisher:       publisher,
		finalizeUseCase: finalizeUseCase,
	}
}

func (d *Dispatcher) HandleAnalyzeRequest(ctx context.Context, message messaging.AnalyzeMessage) error {
	return d.publisher.Publish(ctx, d.queues.MetadataAnalyze, message)
}

func (d *Dispatcher) HandleStageDone(ctx context.Context, message messaging.StageDoneMessage) error {
	media, err := d.mediaRepo.GetByID(ctx, message.MediaID)
	if err != nil {
		return fmt.Errorf("media not found: %w", err)
	}

	next := d.nextStage(message.Stage)
	if next == "" {
		return d.finalizeUseCase.Execute(ctx, media.ID)
	}

	analyzeMessage := messaging.AnalyzeMessage{
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
		return d.queues.MetadataAnalyze
	case "heuristics":
		return d.queues.HeuristicsAnalyze
	case "ai_model":
		return d.queues.AiModelAnalyze
	default:
		return ""
	}
}

func (d *Dispatcher) StageFailedQueue(stage string) string {
	switch stage {
	case "metadata":
		return d.queues.MetadataFailed
	case "heuristics":
		return d.queues.HeuristicsFailed
	case "ai_model":
		return d.queues.AiModelFailed
	default:
		return ""
	}
}

func (d *Dispatcher) StageDoneQueue(stage string) string {
	switch stage {
	case "metadata":
		return d.queues.MetadataDone
	case "heuristics":
		return d.queues.HeuristicsDone
	case "ai_model":
		return d.queues.AiModelDone
	default:
		return ""
	}
}

func (d *Dispatcher) PublishStageDone(ctx context.Context, queueName string, message messaging.StageDoneMessage) error {
	return d.publisher.Publish(ctx, queueName, message)
}

func (d *Dispatcher) PublishFailed(ctx context.Context, queueName string, message messaging.FailedMessage) error {
	return d.publisher.Publish(ctx, queueName, message)
}
