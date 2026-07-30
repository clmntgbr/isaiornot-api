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
	"go-api/usecase/media"
	scanUseCase "go-api/usecase/scan"
	"go-api/usecase/subscription"

	"github.com/google/uuid"
)

type AggregateScanUseCase struct {
	mediaRepo                    repository.MediaRepository
	scanRepo                     repository.ScanRepository
	signalRepo                   repository.SignalRepository
	updateScanStatusUseCase      *scanUseCase.UpdateScanStatusUseCase
	centrifugoPublisher          port.RealtimePublisher
	resolvePipelineAccessUseCase *subscription.ResolvePipelineAccessUseCase
	deleteOriginalsUseCase       *media.DeleteOriginalsUseCase
}

func NewAggregateScanUseCase(
	mediaRepo repository.MediaRepository,
	scanRepo repository.ScanRepository,
	signalRepo repository.SignalRepository,
	updateScanStatusUseCase *scanUseCase.UpdateScanStatusUseCase,
	centrifugoPublisher port.RealtimePublisher,
	resolvePipelineAccessUseCase *subscription.ResolvePipelineAccessUseCase,
	deleteOriginalsUseCase *media.DeleteOriginalsUseCase,
) *AggregateScanUseCase {
	return &AggregateScanUseCase{
		mediaRepo:                    mediaRepo,
		scanRepo:                     scanRepo,
		signalRepo:                   signalRepo,
		updateScanStatusUseCase:      updateScanStatusUseCase,
		centrifugoPublisher:          centrifugoPublisher,
		resolvePipelineAccessUseCase: resolvePipelineAccessUseCase,
		deleteOriginalsUseCase:       deleteOriginalsUseCase,
	}
}

func (u *AggregateScanUseCase) Execute(ctx context.Context, mediaID uuid.UUID) error {
	media, err := u.mediaRepo.GetByID(ctx, mediaID)
	if err != nil {
		return errors.New("media not found")
	}

	fullPipeline, err := u.resolvePipelineAccessUseCase.FullPipelineForUser(ctx, media.UserID)
	if err != nil {
		return err
	}
	required := requiredSignalNames(fullPipeline)

	signals, err := u.signalRepo.GetByMediaID(ctx, mediaID)
	if err != nil {
		return errors.New("failed to load signals")
	}

	if !hasAllRequiredSignals(signals, required) {
		return errors.New("not all signals are ready")
	}

	media.Statuses = append(media.Statuses, enum.MediaStatusCompleted)
	media.Status = enum.MediaStatusCompleted
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

	if scan.Status != enum.ScanStatusCompleted {
		return nil
	}

	mediaResults := make([]aggregate.AggregationResult, 0, len(scan.Medias))
	var allSignals []*entity.Signal

	for _, scanMedia := range scan.Medias {
		mediaSignals, err := u.signalRepo.GetByMediaID(ctx, scanMedia.ID)
		if err != nil {
			return errors.New("failed to load signals")
		}
		if !hasAllRequiredSignals(mediaSignals, required) {
			return errors.New("not all signals are ready")
		}

		mediaResults = append(mediaResults, aggregate.Compute(toEntitySignals(mediaSignals)))
		allSignals = append(allSignals, mediaSignals...)
	}

	result := aggregate.AggregateMediaResults(mediaResults)

	scan.FinalScore = result.FinalScore
	scan.Confidence = result.Confidence
	scan.Verdict = result.Verdict
	scan.FreezeDuration()
	if err := u.scanRepo.Update(ctx, scan); err != nil {
		return err
	}

	u.deleteOriginalsUseCase.Execute(ctx, scan.UserID, scan.Medias)

	realtimeEvent, err := realtime.NewScanCompletedEvent(scan, media, allSignals)
	if err != nil {
		return errors.New("failed to build scan completed event")
	}

	if err := u.centrifugoPublisher.PublishToUser(ctx, scan.UserID, realtimeEvent); err != nil {
		return errors.New("failed to publish scan completed event")
	}

	return nil
}

func requiredSignalNames(fullPipeline bool) []string {
	names := []string{"metadata", "heuristics"}
	if fullPipeline {
		names = append(names, "ai_model")
	}
	return names
}

func hasAllRequiredSignals(signals []*entity.Signal, required []string) bool {
	found := make(map[string]struct{}, len(required))
	for _, signal := range signals {
		found[signal.Name] = struct{}{}
	}

	for _, name := range required {
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
