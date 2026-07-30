package scan

import (
	"context"
	"fmt"
	"time"

	"go-api/domain/entity"
	"go-api/domain/enum"
	"go-api/domain/repository"

	"github.com/google/uuid"
)

const (
	StaleRetryAfter             = time.Hour
	MaxStaleRetries             = 3
	StaleScanTimeoutMessage     = "Scan processing timed out after retries"
	StaleUploadedTimeoutMessage = "Scan stuck in uploaded status"
)

type AnalyzePublisher interface {
	Execute(ctx context.Context, mediaID uuid.UUID) error
}

type RetryStaleScansUseCase struct {
	scanRepo         repository.ScanRepository
	mediaRepo        repository.MediaRepository
	signalRepo       repository.SignalRepository
	analyzePublisher AnalyzePublisher
	failScanUseCase  *FailScanUseCase
}

func NewRetryStaleScansUseCase(
	scanRepo repository.ScanRepository,
	mediaRepo repository.MediaRepository,
	signalRepo repository.SignalRepository,
	analyzePublisher AnalyzePublisher,
	failScanUseCase *FailScanUseCase,
) *RetryStaleScansUseCase {
	return &RetryStaleScansUseCase{
		scanRepo:         scanRepo,
		mediaRepo:        mediaRepo,
		signalRepo:       signalRepo,
		analyzePublisher: analyzePublisher,
		failScanUseCase:  failScanUseCase,
	}
}

type RetryStaleScansResult struct {
	Retried int
	Failed  int
	Medias  int
}

func (u *RetryStaleScansUseCase) Execute(ctx context.Context) (*RetryStaleScansResult, error) {
	before := time.Now().UTC().Add(-StaleRetryAfter)
	scans, err := u.scanRepo.ListInProgressCreatedBefore(ctx, before)
	if err != nil {
		return nil, fmt.Errorf("failed to list stale scans: %w", err)
	}

	result := &RetryStaleScansResult{}
	for _, scanEntity := range scans {
		switch scanEntity.Status {
		case enum.ScanStatusUploaded:
			if err := u.failScanUseCase.Execute(ctx, scanEntity.ID, StaleUploadedTimeoutMessage); err != nil {
				return result, fmt.Errorf("failed to fail uploaded scan %s: %w", scanEntity.ID, err)
			}
			result.Failed++

		case enum.ScanStatusProcessing:
			if err := u.retryOrFailProcessing(ctx, scanEntity, result); err != nil {
				return result, err
			}
		}
	}

	return result, nil
}

func (u *RetryStaleScansUseCase) retryOrFailProcessing(
	ctx context.Context,
	scanEntity *entity.Scan,
	result *RetryStaleScansResult,
) error {
	requiredAge := time.Duration(scanEntity.RetryCount+1) * StaleRetryAfter
	if time.Since(scanEntity.CreatedAt) < requiredAge {
		return nil
	}

	if scanEntity.RetryCount >= MaxStaleRetries {
		if err := u.failScanUseCase.Execute(ctx, scanEntity.ID, StaleScanTimeoutMessage); err != nil {
			return fmt.Errorf("failed to fail processing scan %s: %w", scanEntity.ID, err)
		}
		result.Failed++
		return nil
	}

	retriedMedia := 0
	for i := range scanEntity.Medias {
		mediaEntity := &scanEntity.Medias[i]
		if mediaEntity.Status == enum.MediaStatusCompleted || mediaEntity.Status == enum.MediaStatusFailed {
			continue
		}

		if err := u.signalRepo.DeleteByMediaID(ctx, mediaEntity.ID); err != nil {
			return fmt.Errorf("failed to clear signals for media %s: %w", mediaEntity.ID, err)
		}

		if mediaEntity.Status != enum.MediaStatusUploaded {
			mediaEntity.Statuses = append(mediaEntity.Statuses, enum.MediaStatusUploaded)
			mediaEntity.Status = enum.MediaStatusUploaded
			if err := u.mediaRepo.Update(ctx, mediaEntity); err != nil {
				return fmt.Errorf("failed to reset media %s: %w", mediaEntity.ID, err)
			}
		}

		if err := u.analyzePublisher.Execute(ctx, mediaEntity.ID); err != nil {
			return fmt.Errorf("failed to republish media %s: %w", mediaEntity.ID, err)
		}
		retriedMedia++
	}

	if retriedMedia == 0 {
		return nil
	}

	scanEntity.RetryCount++
	if err := u.scanRepo.Update(ctx, scanEntity); err != nil {
		return fmt.Errorf("failed to update retry count for scan %s: %w", scanEntity.ID, err)
	}

	result.Medias += retriedMedia
	result.Retried++
	return nil
}
