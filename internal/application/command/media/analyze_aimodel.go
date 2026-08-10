package media

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"time"

	signalcmd "go-api/internal/application/command/signal"
	domainmedia "go-api/internal/domain/media"
	"go-api/internal/domain/port"
	domainscan "go-api/internal/domain/scan"

	"github.com/google/uuid"
)

const maxAIModelScanBytes = 8 * 1024 * 1024

var (
	ErrAIModelMediaNotFound = errors.New("media not found")
	ErrAIModelScanNotFound  = errors.New("scan not found")
)

type AnalyzeAIModelCommand struct {
	MediaID uuid.UUID
}

type AnalyzeAIModelHandler struct {
	mediaRepo    domainmedia.MediaWriteRepository
	scanRepo     domainscan.ScanWriteRepository
	storage      port.Storage
	providers    []port.AIModelProvider
	upsertSignal *signalcmd.UpsertSignalHandler
	publisher    port.EventPublisher
}

func NewAnalyzeAIModelHandler(
	mediaRepo domainmedia.MediaWriteRepository,
	scanRepo domainscan.ScanWriteRepository,
	storage port.Storage,
	providers []port.AIModelProvider,
	upsertSignal *signalcmd.UpsertSignalHandler,
	publisher port.EventPublisher,
) *AnalyzeAIModelHandler {
	return &AnalyzeAIModelHandler{
		mediaRepo:    mediaRepo,
		scanRepo:     scanRepo,
		storage:      storage,
		providers:    providers,
		upsertSignal: upsertSignal,
		publisher:    publisher,
	}
}

func (h *AnalyzeAIModelHandler) Handle(ctx context.Context, cmd AnalyzeAIModelCommand) error {
	mediaEntity, err := h.mediaRepo.GetByID(ctx, cmd.MediaID)
	if err != nil {
		return err
	}
	if mediaEntity == nil {
		return ErrAIModelMediaNotFound
	}

	contentType := domainmedia.ContentTypeFromKey(mediaEntity.Key, mediaEntity.ContentType)
	if domainmedia.IsImageContentType(contentType) {
		scanEntity, err := h.scanRepo.GetByID(ctx, mediaEntity.ScanID)
		if err != nil {
			return err
		}
		if scanEntity == nil {
			return ErrAIModelScanNotFound
		}

		objectKey := domainmedia.NewObjectKey(scanEntity.UserID, mediaEntity.ScanID, mediaEntity.Key)
		reader, err := h.storage.Get(ctx, objectKey)
		if err != nil {
			return fmt.Errorf("failed to download media %q: %w", objectKey, err)
		}
		defer reader.Close()

		imageData, err := io.ReadAll(io.LimitReader(reader, maxAIModelScanBytes))
		if err != nil {
			return fmt.Errorf("failed to read media %q: %w", objectKey, err)
		}

		for _, provider := range h.providers {
			result, err := provider.Analyze(ctx, imageData, mediaEntity.Filename)
			if err != nil {
				return fmt.Errorf("failed to analyze media with model %q: %w", provider.Name(), err)
			}

			if _, err = h.upsertSignal.Handle(ctx, signalcmd.UpsertSignalCommand{
				MediaID:    mediaEntity.ID,
				Name:       provider.Name(),
				Score:      result.Score,
				Confidence: result.Confidence,
				Details:    result.Details,
			}); err != nil {
				return err
			}
		}
	}

	return h.publishDone(ctx, mediaEntity.ID, mediaEntity.ScanID)
}

func (h *AnalyzeAIModelHandler) publishDone(ctx context.Context, mediaID, scanID uuid.UUID) error {
	if h.publisher == nil {
		return nil
	}

	now := time.Now().UTC()
	eventID := uuid.New().String()
	payload, err := json.Marshal(domainmedia.MediaAnalyzeAIModelDone{
		ID:        eventID,
		MediaID:   mediaID.String(),
		ScanID:    scanID.String(),
		Stage:     domainmedia.StageAIModel,
		Timestamp: now,
	})
	if err != nil {
		return err
	}

	return h.publisher.Publish(ctx, port.EventEnvelope{
		EventID:     eventID,
		Type:        domainmedia.EventTypeMediaAnalyzeAIModelDone,
		AggregateID: mediaID.String(),
		OccurredAt:  now.Format(time.RFC3339Nano),
		Payload:     payload,
	})
}
