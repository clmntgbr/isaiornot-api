package aimodel

import (
	"context"
	"errors"
	"fmt"
	"io"
	"path/filepath"

	"go-api/domain/entity"
	"go-api/domain/port"
	mediadto "go-api/domain/media"

	"github.com/google/uuid"
)

type AnalyzeMediaAiModelUseCase struct {
	storage  port.Storage
	analyzer port.AiModelAnalyzer
}

func NewAnalyzeMediaAiModelUseCase(
	storage port.Storage,
	analyzer port.AiModelAnalyzer,
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
) (*entity.Signal, error) {
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

	signal, err := uc.analyzer.Analyze(ctx, imageData, filename)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze media with ai model: %w", err)
	}

	return &signal, nil
}
