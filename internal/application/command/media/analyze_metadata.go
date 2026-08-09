package media

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	signalcmd "go-api/internal/application/command/signal"
	domainmedia "go-api/internal/domain/media"
	"go-api/internal/domain/port"
	domainscan "go-api/internal/domain/scan"
	domainsignal "go-api/internal/domain/signal"

	"github.com/google/uuid"
)

var (
	ErrAnalyzeMediaNotFound = errors.New("media not found")
	ErrAnalyzeScanNotFound  = errors.New("scan not found")
)

type AnalyzeMetadataCommand struct {
	MediaID uuid.UUID
}

type AnalyzeMetadataHandler struct {
	mediaRepo    domainmedia.MediaWriteRepository
	scanRepo     domainscan.ScanWriteRepository
	storage      port.Storage
	analyzer     port.MetadataAnalyzer
	upsertSignal *signalcmd.UpsertSignalHandler
	publisher    port.EventPublisher
}

func NewAnalyzeMetadataHandler(
	mediaRepo domainmedia.MediaWriteRepository,
	scanRepo domainscan.ScanWriteRepository,
	storage port.Storage,
	analyzer port.MetadataAnalyzer,
	upsertSignal *signalcmd.UpsertSignalHandler,
	publisher port.EventPublisher,
) *AnalyzeMetadataHandler {
	return &AnalyzeMetadataHandler{
		mediaRepo:    mediaRepo,
		scanRepo:     scanRepo,
		storage:      storage,
		analyzer:     analyzer,
		upsertSignal: upsertSignal,
		publisher:    publisher,
	}
}

func (h *AnalyzeMetadataHandler) Handle(ctx context.Context, cmd AnalyzeMetadataCommand) error {
	mediaEntity, err := h.mediaRepo.GetByID(ctx, cmd.MediaID)
	if err != nil {
		return err
	}
	if mediaEntity == nil {
		return ErrAnalyzeMediaNotFound
	}

	contentType := domainmedia.ContentTypeFromKey(mediaEntity.Key, mediaEntity.ContentType)
	if domainmedia.IsImageContentType(contentType) {
		scanEntity, err := h.scanRepo.GetByID(ctx, mediaEntity.ScanID)
		if err != nil {
			return err
		}
		if scanEntity == nil {
			return ErrAnalyzeScanNotFound
		}

		objectKey := domainmedia.NewObjectKey(scanEntity.UserID, mediaEntity.ScanID, mediaEntity.Key)
		reader, err := h.storage.Get(ctx, objectKey)
		if err != nil {
			return fmt.Errorf("failed to download media %q: %w", objectKey, err)
		}
		defer reader.Close()

		result, err := h.analyzer.Analyze(reader)
		if err != nil {
			return errors.New("failed to analyze media metadata")
		}

		if _, err = h.upsertSignal.Handle(ctx, signalcmd.UpsertSignalCommand{
			MediaID:    mediaEntity.ID,
			Name:       domainsignal.NameMetadata,
			Score:      result.Score,
			Confidence: result.Confidence,
			Details:    result.Details,
		}); err != nil {
			return err
		}
	}

	return h.publishDone(ctx, mediaEntity.ID, mediaEntity.ScanID)
}

func (h *AnalyzeMetadataHandler) publishDone(ctx context.Context, mediaID, scanID uuid.UUID) error {
	if h.publisher == nil {
		return nil
	}

	now := time.Now().UTC()
	eventID := uuid.New().String()
	payload, err := json.Marshal(domainmedia.MediaAnalyzeMetadataDone{
		ID:        eventID,
		MediaID:   mediaID.String(),
		ScanID:    scanID.String(),
		Stage:     domainmedia.StageMetadata,
		Timestamp: now,
	})
	if err != nil {
		return err
	}

	return h.publisher.Publish(ctx, port.EventEnvelope{
		EventID:     eventID,
		Type:        domainmedia.EventTypeMediaAnalyzeMetadataDone,
		AggregateID: mediaID.String(),
		OccurredAt:  now.Format(time.RFC3339Nano),
		Payload:     payload,
	})
}
