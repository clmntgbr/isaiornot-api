package port

import (
	"io"

	domainsignal "go-api/internal/domain/signal"
)

type MetadataAnalysis struct {
	Score      int
	Confidence domainsignal.ConfidenceLevel
	Details    []string
}

type MetadataAnalyzer interface {
	Analyze(r io.Reader) (MetadataAnalysis, error)
}
