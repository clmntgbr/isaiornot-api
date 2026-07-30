package scan

import (
	"context"
	"fmt"
	"time"

	"go-api/domain/enum"
	"go-api/domain/repository"

	"github.com/google/uuid"
)

type AnalyzePublisher interface {
	Execute(ctx context.Context, mediaID uuid.UUID) error
}

type RetryStaleScansUseCase struct {
	scanRepo         repository.ScanRepository
	mediaRepo        repository.MediaRepository
	signalRepo       repository.SignalRepository
	analyzePublisher AnalyzePublisher
}

func NewRetryStaleScansUseCase(
	scanRepo repository.ScanRepository,
	mediaRepo repository.MediaRepository,
	signalRepo repository.SignalRepository,
	analyzePublisher AnalyzePublisher,
) *RetryStaleScansUseCase {
	return &RetryStaleScansUseCase{
		scanRepo:         scanRepo,
		mediaRepo:        mediaRepo,
		signalRepo:       signalRepo,
		analyzePublisher: analyzePublisher,
	}
}

type RetryStaleScansResult struct {
	Retried int
	Medias  int
}

func (u *RetryStaleScansUseCase) Execute(ctx context.Context, olderThan time.Duration) (*RetryStaleScansResult, error) {
	if olderThan <= 0 {
		return nil, fmt.Errorf("olderThan must be positive")
	}

	before := time.Now().UTC().Add(-olderThan)
	scans, err := u.scanRepo.ListInProgressCreatedBefore(ctx, before)
	if err != nil {
		return nil, fmt.Errorf("failed to list stale scans: %w", err)
	}

	result := &RetryStaleScansResult{}
	for _, scan := range scans {
		for i := range scan.Medias {
			mediaEntity := &scan.Medias[i]
			if mediaEntity.Status == enum.MediaStatusCompleted || mediaEntity.Status == enum.MediaStatusFailed {
				continue
			}

			if err := u.signalRepo.DeleteByMediaID(ctx, mediaEntity.ID); err != nil {
				return result, fmt.Errorf("failed to clear signals for media %s: %w", mediaEntity.ID, err)
			}

			if mediaEntity.Status != enum.MediaStatusUploaded {
				mediaEntity.Statuses = append(mediaEntity.Statuses, enum.MediaStatusUploaded)
				mediaEntity.Status = enum.MediaStatusUploaded
				if err := u.mediaRepo.Update(ctx, mediaEntity); err != nil {
					return result, fmt.Errorf("failed to reset media %s: %w", mediaEntity.ID, err)
				}
			}

			if err := u.analyzePublisher.Execute(ctx, mediaEntity.ID); err != nil {
				return result, fmt.Errorf("failed to republish media %s: %w", mediaEntity.ID, err)
			}
			result.Medias++
		}
		result.Retried++
	}

	return result, nil
}
