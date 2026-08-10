package scan

import (
	"context"
	"fmt"
	"time"

	"go-api/internal/domain/event"
	domainmedia "go-api/internal/domain/media"
	"go-api/internal/domain/port"
	domainscan "go-api/internal/domain/scan"
	domainsignal "go-api/internal/domain/signal"
)

const (
	StaleRetryAfter             = 30 * time.Minute
	MaxStaleRetries             = 3
	StaleScanTimeoutMessage     = "Scan processing timed out after retries"
	StaleUploadedTimeoutMessage = "Scan stuck in uploaded status"
)

type RetryStaleScansResult struct {
	Retried int
	Failed  int
	Medias  int
}

type RetryStaleScansHandler struct {
	scanRepo     domainscan.ScanWriteRepository
	mediaRepo    domainmedia.MediaWriteRepository
	signalRepo   domainsignal.SignalWriteRepository
	outbox       port.OutboxRepository
	failScan     *FailScanHandler
}

func NewRetryStaleScansHandler(
	scanRepo domainscan.ScanWriteRepository,
	mediaRepo domainmedia.MediaWriteRepository,
	signalRepo domainsignal.SignalWriteRepository,
	outbox port.OutboxRepository,
	failScan *FailScanHandler,
) *RetryStaleScansHandler {
	return &RetryStaleScansHandler{
		scanRepo:   scanRepo,
		mediaRepo:  mediaRepo,
		signalRepo: signalRepo,
		outbox:     outbox,
		failScan:   failScan,
	}
}

func (h *RetryStaleScansHandler) Handle(ctx context.Context) (*RetryStaleScansResult, error) {
	before := time.Now().UTC().Add(-StaleRetryAfter)
	scans, err := h.scanRepo.ListInProgressCreatedBefore(ctx, before)
	if err != nil {
		return nil, fmt.Errorf("failed to list stale scans: %w", err)
	}

	result := &RetryStaleScansResult{}
	for _, scanEntity := range scans {
		switch scanEntity.Status {
		case domainscan.StatusUploaded:
			if err := h.failScan.Handle(ctx, FailScanCommand{
				ScanID:  scanEntity.ID,
				Message: StaleUploadedTimeoutMessage,
			}); err != nil {
				return result, fmt.Errorf("failed to fail uploaded scan %s: %w", scanEntity.ID, err)
			}
			result.Failed++

		case domainscan.StatusProcessing:
			if err := h.retryOrFailProcessing(ctx, scanEntity, result); err != nil {
				return result, err
			}
		}
	}

	return result, nil
}

func (h *RetryStaleScansHandler) retryOrFailProcessing(
	ctx context.Context,
	scanEntity *domainscan.Scan,
	result *RetryStaleScansResult,
) error {
	requiredAge := time.Duration(scanEntity.RetryCount+1) * StaleRetryAfter
	if time.Since(scanEntity.CreatedAt) < requiredAge {
		return nil
	}

	if scanEntity.RetryCount >= MaxStaleRetries {
		if err := h.failScan.Handle(ctx, FailScanCommand{
			ScanID:  scanEntity.ID,
			Message: StaleScanTimeoutMessage,
		}); err != nil {
			return fmt.Errorf("failed to fail processing scan %s: %w", scanEntity.ID, err)
		}
		result.Failed++
		return nil
	}

	retriedMedia := 0
	err := h.scanRepo.WithTransaction(ctx, func(txCtx context.Context) error {
		freshScan, err := h.scanRepo.GetByID(txCtx, scanEntity.ID)
		if err != nil {
			return err
		}
		if freshScan == nil || freshScan.Status != domainscan.StatusProcessing {
			return nil
		}

		medias, err := h.mediaRepo.GetByScanID(txCtx, freshScan.ID)
		if err != nil {
			return fmt.Errorf("failed to get medias for scan %s: %w", freshScan.ID, err)
		}

		events := make([]event.DomainEvent, 0)
		count := 0
		for _, mediaEntity := range medias {
			if mediaEntity.Status == domainmedia.StatusCompleted || mediaEntity.Status == domainmedia.StatusFailed {
				continue
			}

			if err := h.signalRepo.DeleteByMediaID(txCtx, mediaEntity.ID); err != nil {
				return fmt.Errorf("failed to clear signals for media %s: %w", mediaEntity.ID, err)
			}

			mediaEntity.ResetToUploadedForRetry()
			if err := h.mediaRepo.Update(txCtx, mediaEntity); err != nil {
				return fmt.Errorf("failed to reset media %s: %w", mediaEntity.ID, err)
			}
			events = append(events, mediaEntity.PullEvents()...)
			count++
		}

		if count == 0 {
			return nil
		}

		freshScan.IncrementRetry()
		if err := h.scanRepo.Update(txCtx, freshScan); err != nil {
			return fmt.Errorf("failed to update retry count for scan %s: %w", freshScan.ID, err)
		}
		events = append(events, freshScan.PullEvents()...)

		if err := h.outbox.StoreEvents(txCtx, events); err != nil {
			return err
		}

		retriedMedia = count
		return nil
	})
	if err != nil {
		return err
	}

	if retriedMedia > 0 {
		result.Medias += retriedMedia
		result.Retried++
	}
	return nil
}
