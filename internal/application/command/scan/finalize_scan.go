package scan

import (
	"context"
	"errors"
	"fmt"

	"go-api/internal/domain/aggregate"
	domainmedia "go-api/internal/domain/media"
	"go-api/internal/domain/port"
	domainscan "go-api/internal/domain/scan"
	domainsignal "go-api/internal/domain/signal"

	"github.com/google/uuid"
)

var (
	ErrFinalizeScanNotFound = errors.New("scan not found")
	ErrFinalizeScanNotReady = errors.New("scan medias are not all completed")
)

type FinalizeScanCommand struct {
	ScanID uuid.UUID
}

type FinalizeScanHandler struct {
	scanRepo   domainscan.ScanWriteRepository
	mediaRepo  domainmedia.MediaWriteRepository
	signalRepo domainsignal.SignalWriteRepository
	outbox     port.OutboxRepository
}

func NewFinalizeScanHandler(
	scanRepo domainscan.ScanWriteRepository,
	mediaRepo domainmedia.MediaWriteRepository,
	signalRepo domainsignal.SignalWriteRepository,
	outbox port.OutboxRepository,
) *FinalizeScanHandler {
	return &FinalizeScanHandler{
		scanRepo:   scanRepo,
		mediaRepo:  mediaRepo,
		signalRepo: signalRepo,
		outbox:     outbox,
	}
}

func (h *FinalizeScanHandler) Handle(ctx context.Context, cmd FinalizeScanCommand) error {
	scanEntity, err := h.scanRepo.GetByID(ctx, cmd.ScanID)
	if err != nil {
		return err
	}
	if scanEntity == nil {
		return ErrFinalizeScanNotFound
	}
	if scanEntity.Status == domainscan.StatusCompleted || scanEntity.Status == domainscan.StatusFailed {
		return nil
	}

	medias, err := h.mediaRepo.GetByScanID(ctx, cmd.ScanID)
	if err != nil {
		return err
	}
	if len(medias) == 0 {
		return ErrFinalizeScanNotReady
	}
	for _, media := range medias {
		if media.Status != domainmedia.StatusCompleted {
			return ErrFinalizeScanNotReady
		}
	}

	mediaResults := make([]aggregate.Result, 0, len(medias))
	for _, media := range medias {
		signals, err := h.signalRepo.GetByMediaID(ctx, media.ID)
		if err != nil {
			return err
		}
		inputs := make([]aggregate.InputSignal, 0, len(signals))
		for _, signal := range signals {
			inputs = append(inputs, aggregate.InputSignal{
				Name:       signal.Name,
				Score:      signal.Score,
				Confidence: signal.Confidence,
			})
		}
		mediaResults = append(mediaResults, aggregate.Compute(inputs))
	}

	result := aggregate.AggregateMediaResults(mediaResults)

	err = h.scanRepo.WithTransaction(ctx, func(txCtx context.Context) error {
		fresh, err := h.scanRepo.GetByID(txCtx, cmd.ScanID)
		if err != nil {
			return err
		}
		if fresh == nil {
			return ErrFinalizeScanNotFound
		}
		if fresh.Status == domainscan.StatusCompleted || fresh.Status == domainscan.StatusFailed {
			return nil
		}

		fresh.MarkCompleted(result.FinalScore, result.Confidence, result.Verdict)
		if err := h.scanRepo.Update(txCtx, fresh); err != nil {
			return err
		}
		return h.outbox.StoreEvents(txCtx, fresh.PullEvents())
	})
	if err != nil {
		return fmt.Errorf("failed to finalize scan: %w", err)
	}
	return nil
}
