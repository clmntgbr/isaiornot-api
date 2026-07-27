package aggregate

import "go-api/domain/entity"

var baseWeights = map[string]float64{
	"metadata":   0.20,
	"heuristics": 0.30,
	"ai_model":   0.50,
}

func effectiveWeight(signal entity.Signal) float64 {
	base, ok := baseWeights[signal.Name]
	if !ok {
		return 0
	}

	weight := base * confidenceMultiplier(signal.Confidence)
	if signal.Name == "ai_model" && signal.Confidence == entity.ConfidenceHigh {
		weight *= 1.5
	}

	return weight
}

func confidenceMultiplier(confidence entity.ConfidenceLevel) float64 {
	switch confidence {
	case entity.ConfidenceHigh:
		return 1.0
	case entity.ConfidenceMedium:
		return 0.75
	case entity.ConfidenceLow:
		return 0.5
	default:
		return 0.25
	}
}
