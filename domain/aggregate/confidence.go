package aggregate

import (
	"math"

	"go-api/domain/entity"
)

func agreementConfidence(scores []float64) entity.ConfidenceLevel {
	if len(scores) == 0 {
		return entity.ConfidenceUnknown
	}
	if len(scores) == 1 {
		return entity.ConfidenceHigh
	}

	_, std := meanAndStd(scores)
	switch {
	case std <= 15:
		return entity.ConfidenceHigh
	case std <= 30:
		return entity.ConfidenceMedium
	default:
		return entity.ConfidenceLow
	}
}

type confidenceRank int

const (
	rankUnknown confidenceRank = iota
	rankLow
	rankMedium
	rankHigh
)

func (r confidenceRank) min(other confidenceRank) confidenceRank {
	if r < other {
		return r
	}
	return other
}

func (r confidenceRank) toLevel() entity.ConfidenceLevel {
	switch r {
	case rankHigh:
		return entity.ConfidenceHigh
	case rankMedium:
		return entity.ConfidenceMedium
	case rankLow:
		return entity.ConfidenceLow
	default:
		return entity.ConfidenceUnknown
	}
}

func confidenceRankOf(level entity.ConfidenceLevel) confidenceRank {
	switch level {
	case entity.ConfidenceHigh:
		return rankHigh
	case entity.ConfidenceMedium:
		return rankMedium
	case entity.ConfidenceLow:
		return rankLow
	default:
		return rankUnknown
	}
}

func minConfidenceLevels(levels ...entity.ConfidenceLevel) entity.ConfidenceLevel {
	if len(levels) == 0 {
		return entity.ConfidenceUnknown
	}

	min := confidenceRankOf(levels[0])
	for _, level := range levels[1:] {
		min = min.min(confidenceRankOf(level))
	}
	return min.toLevel()
}

func globalConfidence(available []entity.Signal) entity.ConfidenceLevel {
	count := len(available)
	if count == 0 {
		return entity.ConfidenceUnknown
	}

	scores := make([]float64, 0, count)
	hasHighConfidenceAIModel := false
	for _, signal := range available {
		scores = append(scores, float64(signal.Score))
		if signal.Name == "ai_model" && signal.Confidence == entity.ConfidenceHigh {
			hasHighConfidenceAIModel = true
		}
	}

	_, std := meanAndStd(scores)

	switch {
	case count >= 3 && std <= 20:
		return entity.ConfidenceHigh
	case count >= 3 && std <= 35:
		return entity.ConfidenceMedium
	case hasHighConfidenceAIModel && count >= 2 && std <= 40:
		return entity.ConfidenceMedium
	case count >= 2:
		return entity.ConfidenceLow
	default:
		return entity.ConfidenceLow
	}
}

func meanAndStd(values []float64) (float64, float64) {
	if len(values) == 0 {
		return 0, 0
	}

	var sum float64
	for _, value := range values {
		sum += value
	}

	mean := sum / float64(len(values))
	var variance float64
	for _, value := range values {
		diff := value - mean
		variance += diff * diff
	}

	return mean, math.Sqrt(variance / float64(len(values)))
}
