package sightengine

import (
	"context"
	"fmt"
	"math"
	"sort"

	"go-api/internal/domain/port"
	domainsignal "go-api/internal/domain/signal"
)

type Analyzer struct {
	client *Client
}

func NewAnalyzer(client *Client) *Analyzer {
	return &Analyzer{client: client}
}

func (a *Analyzer) Name() domainsignal.Name {
	return domainsignal.NameSightEngine
}

func (a *Analyzer) Analyze(ctx context.Context, imageData []byte, filename string) (port.AIModelAnalysis, error) {
	response, err := a.client.CheckGenAI(ctx, imageData, filename)
	if err != nil {
		return port.AIModelAnalysis{}, err
	}
	return toAnalysis(*response), nil
}

func toAnalysis(response CheckResponse) port.AIModelAnalysis {
	probability := response.Type.AIGenerated
	score := int(math.Round(probability * 100))

	details := []string{
		fmt.Sprintf("sightengine ai_generated=%.3f", probability),
	}
	details = append(details, topGenerators(response.Type.AIGenerators, 3)...)

	return port.AIModelAnalysis{
		Score:      score,
		Confidence: confidenceFromProbability(probability),
		Details:    details,
	}
}

func confidenceFromProbability(probability float64) domainsignal.ConfidenceLevel {
	switch {
	case probability >= 0.75:
		return domainsignal.ConfidenceHigh
	case probability >= 0.40:
		return domainsignal.ConfidenceMedium
	default:
		return domainsignal.ConfidenceLow
	}
}

func topGenerators(generators map[string]float64, limit int) []string {
	type generatorScore struct {
		name  string
		score float64
	}

	ranked := make([]generatorScore, 0, len(generators))
	for name, score := range generators {
		ranked = append(ranked, generatorScore{name: name, score: score})
	}

	sort.Slice(ranked, func(i, j int) bool {
		return ranked[i].score > ranked[j].score
	})
	if len(ranked) > limit {
		ranked = ranked[:limit]
	}

	details := make([]string, 0, len(ranked))
	for _, item := range ranked {
		details = append(details, fmt.Sprintf("generator %s=%.3f", item.name, item.score))
	}
	return details
}

var _ port.AIModelProvider = (*Analyzer)(nil)
