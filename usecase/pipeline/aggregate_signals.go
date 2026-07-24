package pipeline

import (
	"context"
	"errors"

	"go-api/domain/aggregate"
	"go-api/domain/entity"
	"go-api/domain/enum"
	"go-api/domain/repository"
	"go-api/infrastructure/centrifugo"
	analysisuc "go-api/usecase/analysis"

	"github.com/google/uuid"
)

var requiredSignalNames = []string{"metadata", "heuristics", "ai_model"}

type AggregateAnalysisUseCase struct {
	mediaRepo                   *repository.MediaRepository
	analysisRepo                *repository.AnalysisRepository
	signalRepo                  *repository.SignalRepository
	updateAnalysisStatusUseCase *analysisuc.UpdateAnalysisStatusUseCase
	centrifugoPublisher         *centrifugo.Publisher
}

func NewAggregateAnalysisUseCase(
	mediaRepo *repository.MediaRepository,
	analysisRepo *repository.AnalysisRepository,
	signalRepo *repository.SignalRepository,
	updateAnalysisStatusUseCase *analysisuc.UpdateAnalysisStatusUseCase,
	centrifugoPublisher *centrifugo.Publisher,
) *AggregateAnalysisUseCase {
	return &AggregateAnalysisUseCase{
		mediaRepo:                   mediaRepo,
		analysisRepo:                analysisRepo,
		signalRepo:                  signalRepo,
		updateAnalysisStatusUseCase: updateAnalysisStatusUseCase,
		centrifugoPublisher:         centrifugoPublisher,
	}
}

func (u *AggregateAnalysisUseCase) Execute(ctx context.Context, mediaID uuid.UUID) error {
	media, err := (*u.mediaRepo).GetByID(ctx, mediaID)
	if err != nil {
		return errors.New("media not found")
	}

	signals, err := (*u.signalRepo).GetByMediaID(ctx, mediaID)
	if err != nil {
		return errors.New("failed to load signals")
	}

	if !hasAllRequiredSignals(signals) {
		return errors.New("not all signals are ready")
	}

	media.Statuses = append(media.Statuses, enum.MediaStatusAnalyzed)
	media.Status = enum.MediaStatusAnalyzed
	if err := (*u.mediaRepo).Update(ctx, media); err != nil {
		return err
	}

	if err := u.updateAnalysisStatusUseCase.Execute(ctx, media.AnalysisID); err != nil {
		return err
	}

	analysis, err := (*u.analysisRepo).GetByID(ctx, media.AnalysisID)
	if err != nil {
		return errors.New("analysis not found")
	}

	if analysis.Status != enum.AnalysisStatusAnalyzed {
		return nil
	}

	mediaResults := make([]aggregate.AggregationResult, 0, len(analysis.Medias))
	var allSignals []*entity.Signal

	for _, analysisMedia := range analysis.Medias {
		mediaSignals, err := (*u.signalRepo).GetByMediaID(ctx, analysisMedia.ID)
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

	analysis.FinalScore = result.FinalScore
	analysis.AnalysisConfidence = result.Confidence
	analysis.Verdict = result.Verdict
	if err := (*u.analysisRepo).Update(ctx, analysis); err != nil {
		return err
	}

	realtimeEvent, err := centrifugo.NewAnalysisCompletedEvent(analysis, media, allSignals)
	if err != nil {
		return errors.New("failed to build analysis completed event")
	}

	if err := u.centrifugoPublisher.PublishToUser(ctx, analysis.UserID, realtimeEvent); err != nil {
		return errors.New("failed to publish analysis completed event")
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
