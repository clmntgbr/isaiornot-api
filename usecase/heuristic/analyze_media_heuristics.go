package heuristic

import (
	"context"
	"errors"
	"fmt"

	"go-api/domain/port"
	mediadto "go-api/domain/media"

	"github.com/google/uuid"
)

type AnalyzeMediaHeuristicsUseCase struct {
	storage  port.Storage
	analyzer port.HeuristicsAnalyzer
}

func NewAnalyzeMediaHeuristicsUseCase(
	storage port.Storage,
	analyzer port.HeuristicsAnalyzer,
) *AnalyzeMediaHeuristicsUseCase {
	return &AnalyzeMediaHeuristicsUseCase{
		storage:  storage,
		analyzer: analyzer,
	}
}

func (uc *AnalyzeMediaHeuristicsUseCase) Execute(
	ctx context.Context,
	userID uuid.UUID,
	mediaKey string,
) (*port.HeuristicsScanResult, error) {
	objectKey := mediadto.NewObjectKey(userID, mediaKey)

	reader, err := uc.storage.Get(ctx, objectKey)
	if err != nil {
		return nil, fmt.Errorf("failed to download media %q: %w", objectKey, err)
	}
	defer reader.Close()

	result, err := uc.analyzer.Analyze(reader)
	if err != nil {
		return nil, errors.New("failed to analyze media heuristics")
	}

	return &result, nil
}
