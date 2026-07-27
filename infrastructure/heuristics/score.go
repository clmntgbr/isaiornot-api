package heuristics

import (
	"fmt"
	"math"

	"go-api/domain/entity"
)

const signalName = "heuristics"

const (
	weightNoise       = 0.35
	weightCompression = 0.30
	weightFrequency   = 0.20
	weightHistogram   = 0.15
)

type HeuristicsResult struct {
	NoiseScore       float64 `json:"noise_score"`
	CompressionScore float64 `json:"compression_score"`
	FrequencyScore   float64 `json:"frequency_score"`
	HistogramScore   float64 `json:"histogram_score"`
	Width            int     `json:"width"`
	Height           int     `json:"height"`
	Format           string  `json:"format"`
}

type ScanResult struct {
	Signal     entity.Signal    `json:"signal"`
	Heuristics HeuristicsResult `json:"heuristics"`
}

func (r HeuristicsResult) ToSignal() entity.Signal {
	combined := weightNoise*r.NoiseScore +
		weightCompression*r.CompressionScore +
		weightFrequency*r.FrequencyScore +
		weightHistogram*r.HistogramScore

	score := int(math.Round(clampScore(combined)))
	confidence := r.confidenceLevel()
	details := []string{
		fmt.Sprintf("noise=%.1f", r.NoiseScore),
		fmt.Sprintf("compression=%.1f", r.CompressionScore),
		fmt.Sprintf("frequency=%.1f (weak against modern diffusion models)", r.FrequencyScore),
		fmt.Sprintf("histogram=%.1f", r.HistogramScore),
		"supporting signal only — easily degraded by recompression or resizing",
		"should weigh less than the AI model signal in final aggregation",
	}

	if confidence == entity.ConfidenceLow {
		details = append(details, "low reliability: image too small or heavily degraded for strong heuristics")
	}

	return entity.Signal{
		Name:       signalName,
		Score:      score,
		Confidence: confidence,
		Details:    details,
	}
}

func (r HeuristicsResult) confidenceLevel() entity.ConfidenceLevel {
	pixels := r.Width * r.Height
	if pixels < minReliablePixels {
		return entity.ConfidenceLow
	}

	combined := weightNoise*r.NoiseScore +
		weightCompression*r.CompressionScore +
		weightFrequency*r.FrequencyScore +
		weightHistogram*r.HistogramScore

	if combined >= 65 {
		return entity.ConfidenceMedium
	}

	return entity.ConfidenceLow
}
