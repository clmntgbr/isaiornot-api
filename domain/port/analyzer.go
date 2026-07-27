package port

import (
	"context"
	"io"

	"go-api/domain/entity"
)

type AiModelAnalyzer interface {
	Analyze(ctx context.Context, imageData []byte, filename string) (entity.Signal, error)
}

type MetadataAnalyzer interface {
	Analyze(r io.Reader) (entity.Signal, error)
}

type HeuristicsScanResult struct {
	Signal           entity.Signal
	NoiseScore       float64
	CompressionScore float64
	FrequencyScore   float64
	HistogramScore   float64
}

type HeuristicsAnalyzer interface {
	Analyze(r io.Reader) (HeuristicsScanResult, error)
}
