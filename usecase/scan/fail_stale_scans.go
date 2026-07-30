package scan

import (
	"context"
	"fmt"
	"time"

	"go-api/domain/repository"
)

const StaleScanTimeoutMessage = "Scan processing timed out"

type FailStaleScansUseCase struct {
	scanRepo        repository.ScanRepository
	failScanUseCase *FailScanUseCase
}

func NewFailStaleScansUseCase(
	scanRepo repository.ScanRepository,
	failScanUseCase *FailScanUseCase,
) *FailStaleScansUseCase {
	return &FailStaleScansUseCase{
		scanRepo:        scanRepo,
		failScanUseCase: failScanUseCase,
	}
}

type FailStaleScansResult struct {
	Failed int
}

func (u *FailStaleScansUseCase) Execute(ctx context.Context, olderThan time.Duration) (*FailStaleScansResult, error) {
	if olderThan <= 0 {
		return nil, fmt.Errorf("olderThan must be positive")
	}

	before := time.Now().UTC().Add(-olderThan)
	scans, err := u.scanRepo.ListInProgressCreatedBefore(ctx, before)
	if err != nil {
		return nil, fmt.Errorf("failed to list stale scans: %w", err)
	}

	result := &FailStaleScansResult{}
	for _, scan := range scans {
		if err := u.failScanUseCase.Execute(ctx, scan.ID, StaleScanTimeoutMessage); err != nil {
			return result, fmt.Errorf("failed to fail scan %s: %w", scan.ID, err)
		}
		result.Failed++
	}

	return result, nil
}
