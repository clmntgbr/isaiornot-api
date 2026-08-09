package media

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	signalcmd "go-api/internal/application/command/signal"
	domaininsight "go-api/internal/domain/insight"
	domainmedia "go-api/internal/domain/media"
	"go-api/internal/domain/port"
	domainscan "go-api/internal/domain/scan"
	domainsignal "go-api/internal/domain/signal"

	"github.com/google/uuid"
)

var (
	ErrHeuristicsMediaNotFound = errors.New("media not found")
	ErrHeuristicsScanNotFound  = errors.New("scan not found")
)

type AnalyzeHeuristicsCommand struct {
	MediaID uuid.UUID
}

type AnalyzeHeuristicsHandler struct {
	mediaRepo    domainmedia.MediaWriteRepository
	scanRepo     domainscan.ScanWriteRepository
	insightRepo  domaininsight.InsightWriteRepository
	storage      port.Storage
	analyzer     port.HeuristicsAnalyzer
	upsertSignal *signalcmd.UpsertSignalHandler
	outbox       port.OutboxRepository
	publisher    port.EventPublisher
}

func NewAnalyzeHeuristicsHandler(
	mediaRepo domainmedia.MediaWriteRepository,
	scanRepo domainscan.ScanWriteRepository,
	insightRepo domaininsight.InsightWriteRepository,
	storage port.Storage,
	analyzer port.HeuristicsAnalyzer,
	upsertSignal *signalcmd.UpsertSignalHandler,
	outbox port.OutboxRepository,
	publisher port.EventPublisher,
) *AnalyzeHeuristicsHandler {
	return &AnalyzeHeuristicsHandler{
		mediaRepo:    mediaRepo,
		scanRepo:     scanRepo,
		insightRepo:  insightRepo,
		storage:      storage,
		analyzer:     analyzer,
		upsertSignal: upsertSignal,
		outbox:       outbox,
		publisher:    publisher,
	}
}

func (h *AnalyzeHeuristicsHandler) Handle(ctx context.Context, cmd AnalyzeHeuristicsCommand) error {
	mediaEntity, err := h.mediaRepo.GetByID(ctx, cmd.MediaID)
	if err != nil {
		return err
	}
	if mediaEntity == nil {
		return ErrHeuristicsMediaNotFound
	}

	contentType := domainmedia.ContentTypeFromKey(mediaEntity.Key, mediaEntity.ContentType)
	if !domainmedia.IsImageContentType(contentType) {
		return h.publishDone(ctx, mediaEntity.ID, mediaEntity.ScanID)
	}

	scanEntity, err := h.scanRepo.GetByID(ctx, mediaEntity.ScanID)
	if err != nil {
		return err
	}
	if scanEntity == nil {
		return ErrHeuristicsScanNotFound
	}

	objectKey := domainmedia.NewObjectKey(scanEntity.UserID, mediaEntity.ScanID, mediaEntity.Key)
	reader, err := h.storage.Get(ctx, objectKey)
	if err != nil {
		return fmt.Errorf("failed to download media %q: %w", objectKey, err)
	}
	defer reader.Close()

	result, err := h.analyzer.Analyze(reader)
	if err != nil {
		return fmt.Errorf("failed to analyze media heuristics: %w", err)
	}

	if err := h.persistInsight(ctx, mediaEntity, result); err != nil {
		return err
	}

	if _, err = h.upsertSignal.Handle(ctx, signalcmd.UpsertSignalCommand{
		MediaID:    mediaEntity.ID,
		Name:       domainsignal.NameHeuristics,
		Score:      result.Score,
		Confidence: result.Confidence,
		Details:    result.Details,
	}); err != nil {
		return err
	}

	return h.publishDone(ctx, mediaEntity.ID, mediaEntity.ScanID)
}

func (h *AnalyzeHeuristicsHandler) persistInsight(
	ctx context.Context,
	mediaEntity *domainmedia.Media,
	result port.HeuristicsAnalysis,
) error {
	return h.mediaRepo.WithTransaction(ctx, func(txCtx context.Context) error {
		if mediaEntity.InsightID != nil {
			existing, err := h.insightRepo.GetByID(txCtx, *mediaEntity.InsightID)
			if err != nil {
				return err
			}
			if existing != nil {
				existing.UpdateScores(
					result.NoiseScore,
					result.CompressionScore,
					result.FrequencyScore,
					result.HistogramScore,
				)
				return h.insightRepo.Update(txCtx, existing)
			}
		}

		insight := domaininsight.NewInsight(
			result.NoiseScore,
			result.CompressionScore,
			result.FrequencyScore,
			result.HistogramScore,
		)
		if err := h.insightRepo.Save(txCtx, insight); err != nil {
			return err
		}

		mediaEntity.SetInsightID(insight.ID)
		if err := h.mediaRepo.Update(txCtx, mediaEntity); err != nil {
			return err
		}
		return h.outbox.StoreEvents(txCtx, mediaEntity.PullEvents())
	})
}

func (h *AnalyzeHeuristicsHandler) publishDone(ctx context.Context, mediaID, scanID uuid.UUID) error {
	if h.publisher == nil {
		return nil
	}

	now := time.Now().UTC()
	eventID := uuid.New().String()
	payload, err := json.Marshal(domainmedia.MediaAnalyzeHeuristicsDone{
		ID:        eventID,
		MediaID:   mediaID.String(),
		ScanID:    scanID.String(),
		Stage:     domainmedia.StageHeuristics,
		Timestamp: now,
	})
	if err != nil {
		return err
	}

	return h.publisher.Publish(ctx, port.EventEnvelope{
		EventID:     eventID,
		Type:        domainmedia.EventTypeMediaAnalyzeHeuristicsDone,
		AggregateID: mediaID.String(),
		OccurredAt:  now.Format(time.RFC3339Nano),
		Payload:     payload,
	})
}
