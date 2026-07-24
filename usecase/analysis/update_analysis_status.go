package analysis

import (
	"context"
	"errors"
	"go-api/domain/entity"
	"go-api/domain/enum"
	"go-api/domain/repository"

	"github.com/google/uuid"
)

type UpdateAnalysisStatusUseCase struct {
	analysisRepo *repository.AnalysisRepository
}

func NewUpdateAnalysisStatusUseCase(
	analysisRepo *repository.AnalysisRepository,
) *UpdateAnalysisStatusUseCase {
	return &UpdateAnalysisStatusUseCase{
		analysisRepo: analysisRepo,
	}
}

func (u *UpdateAnalysisStatusUseCase) Execute(ctx context.Context, analysisID uuid.UUID) error {
	analysis, err := (*u.analysisRepo).GetByID(ctx, analysisID)
	if err != nil {
		return errors.New("failed to get analysis")
	}

	status := ResolveStatus(analysis.Medias)
	if analysis.Status == status {
		return nil
	}

	analysis.Statuses = append(analysis.Statuses, status)
	analysis.Status = status

	if err := (*u.analysisRepo).Update(ctx, analysis); err != nil {
		return errors.New("failed to update analysis status")
	}

	return nil
}

func ResolveStatus(medias []entity.Media) enum.AnalysisStatus {
	if len(medias) == 0 {
		return enum.AnalysisStatusProcessing
	}

	allAnalyzed := true
	hasUploaded := false
	for _, media := range medias {
		if media.Status != enum.MediaStatusAnalyzed {
			allAnalyzed = false
		}
		if media.Status == enum.MediaStatusUploaded || media.Status == enum.MediaStatusAnalyzed {
			hasUploaded = true
		}
	}

	if allAnalyzed {
		return enum.AnalysisStatusAnalyzed
	}
	if hasUploaded {
		return enum.AnalysisStatusUploaded
	}

	return enum.AnalysisStatusProcessing
}
