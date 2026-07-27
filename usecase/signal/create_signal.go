package signal

import (
	"context"
	"errors"
	"go-api/domain/entity"
	"go-api/domain/repository"

	"github.com/google/uuid"
)

type CreateSignalUseCase struct {
	signalRepo repository.SignalRepository
}

func NewCreateSignalUseCase(signalRepo repository.SignalRepository) *CreateSignalUseCase {
	return &CreateSignalUseCase{signalRepo: signalRepo}
}

func (u *CreateSignalUseCase) Execute(ctx context.Context, mediaID uuid.UUID, name string, score int, confidence entity.ConfidenceLevel, details []string) (*entity.Signal, error) {
	signal := entity.Signal{
		MediaID:    mediaID,
		Name:       name,
		Score:      score,
		Confidence: confidence,
		Details:    details,
	}

	err := u.signalRepo.Create(ctx, &signal)
	if err != nil {
		return nil, errors.New("failed to create signal")
	}

	return &signal, nil
}
