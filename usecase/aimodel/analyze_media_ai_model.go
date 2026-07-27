package aimodel

import (
	"context"
	"errors"
	"fmt"
	"io"
	"path/filepath"

	aimodelinfra "go-api/infrastructure/aimodel"
	mediadto "go-api/infrastructure/media"
	"go-api/infrastructure/storage"

	"github.com/google/uuid"
)

type AnalyzeMediaAiModelUseCase struct {
	storage  *storage.MinIOStorage
	analyzer *aimodelinfra.Analyzer
}

func NewAnalyzeMediaAiModelUseCase(
	storage *storage.MinIOStorage,
	analyzer *aimodelinfra.Analyzer,
) *AnalyzeMediaAiModelUseCase {
	return &AnalyzeMediaAiModelUseCase{
		storage:  storage,
		analyzer: analyzer,
	}
}

func (uc *AnalyzeMediaAiModelUseCase) Execute(
	ctx context.Context,
	userID uuid.UUID,
	mediaKey string,
) (*aimodelinfra.ScanResult, error) {
	objectKey := mediadto.NewObjectKey(userID, mediaKey)

	reader, err := uc.storage.Get(ctx, objectKey)
	if err != nil {
		return nil, fmt.Errorf("failed to download media %q: %w", objectKey, err)
	}
	defer reader.Close()

	imageData, err := io.ReadAll(io.LimitReader(reader, 8<<20))
	if err != nil {
		return nil, errors.New("failed to read media")
	}

	filename := filepath.Base(mediaKey)
	if filename == "" || filename == "." {
		filename = "media.jpg"
	}

	result, err := uc.analyzer.Analyze(ctx, imageData, filename)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze media with ai model: %w", err)
	}

	return &result, nil
}
