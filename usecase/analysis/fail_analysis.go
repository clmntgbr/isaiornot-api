package analysis

import (
	"context"
	"errors"
	"go-api/domain/enum"
	"go-api/domain/repository"
	"go-api/infrastructure/centrifugo"

	"github.com/google/uuid"
)

type FailAnalysisUseCase struct {
	analysisRepo        *repository.AnalysisRepository
	centrifugoPublisher *centrifugo.Publisher
}

func NewFailAnalysisUseCase(
	analysisRepo *repository.AnalysisRepository,
	centrifugoPublisher *centrifugo.Publisher,
) *FailAnalysisUseCase {
	return &FailAnalysisUseCase{
		analysisRepo:        analysisRepo,
		centrifugoPublisher: centrifugoPublisher,
	}
}

func (u *FailAnalysisUseCase) Execute(ctx context.Context, analysisID uuid.UUID, message string) error {
	analysis, err := (*u.analysisRepo).GetByID(ctx, analysisID)
	if err != nil {
		return errors.New("failed to get analysis")
	}

	analysis.Message = message
	if analysis.Status != enum.AnalysisStatusFailed {
		analysis.Statuses = append(analysis.Statuses, enum.AnalysisStatusFailed)
		analysis.Status = enum.AnalysisStatusFailed
	}

	if err := (*u.analysisRepo).Update(ctx, analysis); err != nil {
		return errors.New("failed to update analysis")
	}

	analysis, err = (*u.analysisRepo).GetByID(ctx, analysisID)
	if err != nil {
		return errors.New("failed to reload analysis")
	}

	event, err := centrifugo.NewAnalysisFailedEvent(analysis)
	if err != nil {
		return errors.New("failed to build analysis failed event")
	}

	if err := u.centrifugoPublisher.PublishToUser(ctx, analysis.UserID, event); err != nil {
		return errors.New("failed to publish analysis failed event")
	}

	return nil
}
