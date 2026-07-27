package aimodel

import (
	"context"
	"fmt"
	"math"
	"sort"

	"go-api/domain/entity"
	"go-api/infrastructure/sightengine"
)

const signalName = "ai_model"

type ScanResult struct {
	Signal      entity.Signal             `json:"signal"`
	Sightengine sightengine.CheckResponse `json:"sightengine"`
}

type Analyzer struct {
	client *sightengine.Client
}

func NewAnalyzer(client *sightengine.Client) *Analyzer {
	return &Analyzer{client: client}
}

func (a *Analyzer) Analyze(ctx context.Context, imageData []byte, filename string) (ScanResult, error) {
	response, err := a.client.CheckGenAI(ctx, imageData, filename)
	if err != nil {
		return ScanResult{}, err
	}

	signal := toSignal(*response)

	return ScanResult{
		Signal:      signal,
		Sightengine: *response,
	}, nil
}

func toSignal(response sightengine.CheckResponse) entity.Signal {
	probability := response.Type.AIGenerated
	score := int(math.Round(probability * 100))

	details := []string{
		fmt.Sprintf("sightengine ai_generated=%.3f", probability),
	}
	details = append(details, topGenerators(response.Type.AIGenerators, 3)...)

	return entity.Signal{
		Name:       signalName,
		Score:      score,
		Confidence: confidenceFromProbability(probability),
		Details:    details,
	}
}

func confidenceFromProbability(probability float64) entity.ConfidenceLevel {
	switch {
	case probability >= 0.75:
		return entity.ConfidenceHigh
	case probability >= 0.40:
		return entity.ConfidenceMedium
	default:
		return entity.ConfidenceLow
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

	details := make([]string, 0, limit)
	for _, item := range ranked {
		details = append(details, fmt.Sprintf("generator %s=%.3f", item.name, item.score))
	}

	return details
}
