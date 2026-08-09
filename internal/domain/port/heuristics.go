package port

import (
	"io"

	domainsignal "go-api/internal/domain/signal"
)

type HeuristicsAnalysis struct {
	Score            int
	Confidence       domainsignal.ConfidenceLevel
	Details          []string
	NoiseScore       float64
	CompressionScore float64
	FrequencyScore   float64
	HistogramScore   float64
}

type HeuristicsAnalyzer interface {
	Analyze(r io.Reader) (HeuristicsAnalysis, error)
}
