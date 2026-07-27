package metadata

import (
	"context"
	"errors"
	"fmt"

	mediadto "go-api/infrastructure/media"
	metadatainfra "go-api/infrastructure/metadata"
	"go-api/infrastructure/storage"

	"github.com/google/uuid"
)

type AnalyzeMediaMetadataUseCase struct {
	storage  *storage.MinIOStorage
	analyzer *metadatainfra.Analyzer
}

func NewAnalyzeMediaMetadataUseCase(
	storage *storage.MinIOStorage,
	analyzer *metadatainfra.Analyzer,
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
) (*metadatainfra.ScanResult, error) {
	objectKey := mediadto.NewObjectKey(userID, mediaKey)

	reader, err := uc.storage.Get(ctx, objectKey)
	if err != nil {
		return nil, fmt.Errorf("failed to download media %q: %w", objectKey, err)
	}
	defer reader.Close()

	result, err := uc.analyzer.Analyze(reader)
	if err != nil {
		return nil, errors.New("failed to analyze media metadata")
	}

	return &result, nil
}
