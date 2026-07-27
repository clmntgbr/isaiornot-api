package insight

import (
	"context"
	"errors"
	"go-api/domain/entity"
	"go-api/domain/repository"

	"github.com/google/uuid"
)

type CreateInsightUseCase struct {
	insightRepo repository.InsightRepository
	mediaRepo   repository.MediaRepository
}

func NewCreateInsightUseCase(
	insightRepo repository.InsightRepository,
	mediaRepo repository.MediaRepository,
) *CreateInsightUseCase {
	return &CreateInsightUseCase{
		insightRepo: insightRepo,
		mediaRepo:   mediaRepo,
	}
}

func (u *CreateInsightUseCase) Execute(
	ctx context.Context,
	mediaID uuid.UUID,
	noise float64,
	compression float64,
	frequency float64,
	histogram float64,
) (*entity.Insight, error) {
	media, err := u.mediaRepo.GetByID(ctx, mediaID)
	if err != nil {
		return nil, errors.New("failed to get media")
	}

	insight := entity.Insight{
		Noise:       noise,
		Compression: compression,
		Frequency:   frequency,
		Histogram:   histogram,
	}

	if err := u.insightRepo.Create(ctx, &insight); err != nil {
		return nil, errors.New("failed to create insight")
	}

	media.InsightID = &insight.ID
	if err := u.mediaRepo.Update(ctx, media); err != nil {
		return nil, errors.New("failed to link insight to media")
	}

	return &insight, nil
}
