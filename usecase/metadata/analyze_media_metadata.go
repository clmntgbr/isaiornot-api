package metadata

import (
	"context"
	"errors"
	"fmt"

	"go-api/domain/entity"
	"go-api/domain/port"
	mediadto "go-api/domain/media"

	"github.com/google/uuid"
)

type AnalyzeMediaMetadataUseCase struct {
	storage  port.Storage
	analyzer port.MetadataAnalyzer
}

func NewAnalyzeMediaMetadataUseCase(
	storage port.Storage,
	analyzer port.MetadataAnalyzer,
) *AnalyzeMediaMetadataUseCase {
	return &AnalyzeMediaMetadataUseCase{
		storage:  storage,
		analyzer: analyzer,
	}
}

func (uc *AnalyzeMediaMetadataUseCase) Execute(
	ctx context.Context,
	userID uuid.UUID,
	mediaKey string,
) (*entity.Signal, error) {
	objectKey := mediadto.NewObjectKey(userID, mediaKey)

	reader, err := uc.storage.Get(ctx, objectKey)
	if err != nil {
		return nil, fmt.Errorf("failed to download media %q: %w", objectKey, err)
	}
	defer reader.Close()

	signal, err := uc.analyzer.Analyze(reader)
	if err != nil {
		return nil, errors.New("failed to analyze media metadata")
	}

	return &signal, nil
}
