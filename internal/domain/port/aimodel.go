package port

import (
	"context"

	domainsignal "go-api/internal/domain/signal"
)

type AIModelAnalysis struct {
	Score      int
	Confidence domainsignal.ConfidenceLevel
	Details    []string
}

type AIModelProvider interface {
	Name() domainsignal.Name
	Analyze(ctx context.Context, imageData []byte, filename string) (AIModelAnalysis, error)
}
