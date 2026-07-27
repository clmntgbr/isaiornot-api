package pipeline

import (
	"context"
	"errors"

	"go-api/domain/aggregate"
	"go-api/domain/entity"
	"go-api/domain/enum"
	"go-api/domain/port"
	"go-api/domain/realtime"
	"go-api/domain/repository"
	scanuc "go-api/usecase/scan"

	"github.com/google/uuid"
)

var requiredSignalNames = []string{"metadata", "heuristics", "ai_model"}

type AggregateScanUseCase struct {
	mediaRepo               repository.MediaRepository
	scanRepo                repository.ScanRepository
	signalRepo              repository.SignalRepository
	updateScanStatusUseCase *scanuc.UpdateScanStatusUseCase
	centrifugoPublisher     port.RealtimePublisher
}

func NewAggregateScanUseCase(
	mediaRepo repository.MediaRepository,
	scanRepo repository.ScanRepository,
	signalRepo repository.SignalRepository,
	updateScanStatusUseCase *scanuc.UpdateScanStatusUseCase,
	centrifugoPublisher port.RealtimePublisher,
) *AggregateScanUseCase {
	return &AggregateScanUseCase{
		mediaRepo:               mediaRepo,
		scanRepo:                scanRepo,
		signalRepo:              signalRepo,
		updateScanStatusUseCase: updateScanStatusUseCase,
		centrifugoPublisher:     centrifugoPublisher,
	}
}

func (u *AggregateScanUseCase) Execute(ctx context.Context, mediaID uuid.UUID) error {
	media, err := u.mediaRepo.GetByID(ctx, mediaID)
	if err != nil {
		return errors.New("media not found")
	}

	signals, err := u.signalRepo.GetByMediaID(ctx, mediaID)
	if err != nil {
		return errors.New("failed to load signals")
	}

	if !hasAllRequiredSignals(signals) {
		return errors.New("not all signals are ready")
	}

	media.Statuses = append(media.Statuses, enum.MediaStatusAnalyzed)
	media.Status = enum.MediaStatusAnalyzed
	if err := u.mediaRepo.Update(ctx, media); err != nil {
		return err
	}

	if err := u.updateScanStatusUseCase.Execute(ctx, media.ScanID); err != nil {
		return err
	}

	scan, err := u.scanRepo.GetByID(ctx, media.ScanID)
	if err != nil {
		return errors.New("scan not found")
	}

	if scan.Status != enum.ScanStatusAnalyzed {
		return nil
	}

	mediaResults := make([]aggregate.AggregationResult, 0, len(scan.Medias))
	var allSignals []*entity.Signal

	for _, scanMedia := range scan.Medias {
		mediaSignals, err := u.signalRepo.GetByMediaID(ctx, scanMedia.ID)
		if err != nil {
			return errors.New("failed to load signals")
		}
		if !hasAllRequiredSignals(mediaSignals) {
			return errors.New("not all signals are ready")
		}

		mediaResults = append(mediaResults, aggregate.Compute(toEntitySignals(mediaSignals)))
		allSignals = append(allSignals, mediaSignals...)
	}

	result := aggregate.AggregateMediaResults(mediaResults)

	scan.FinalScore = result.FinalScore
	scan.ScanConfidence = result.Confidence
	scan.Verdict = result.Verdict
	if err := u.scanRepo.Update(ctx, scan); err != nil {
		return err
	}

	realtimeEvent, err := realtime.NewScanCompletedEvent(scan, media, allSignals)
	if err != nil {
		return errors.New("failed to build scan completed event")
	}

	if err := u.centrifugoPublisher.PublishToUser(ctx, scan.UserID, realtimeEvent); err != nil {
		return errors.New("failed to publish scan completed event")
	}

	return nil
}

func hasAllRequiredSignals(signals []*entity.Signal) bool {
	found := make(map[string]struct{}, len(requiredSignalNames))
	for _, signal := range signals {
		found[signal.Name] = struct{}{}
	}

	for _, name := range requiredSignalNames {
		if _, ok := found[name]; !ok {
			return false
		}
	}

	return true
}

func toEntitySignals(signals []*entity.Signal) []entity.Signal {
	result := make([]entity.Signal, 0, len(signals))
	for _, signal := range signals {
		if signal == nil {
			continue
		}
		result = append(result, *signal)
	}

	return result
}
