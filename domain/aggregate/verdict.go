package aggregate

import (
	"math"

	"go-api/domain/entity"
)

const (
	VerdictLikelyReal = "likely_real"
	VerdictUncertain  = "uncertain"
	VerdictLikelyAI   = "likely_ai"
)

type AggregationResult struct {
	FinalScore float64
	Confidence entity.ConfidenceLevel
	Verdict    string
	Signals    []entity.Signal
}

var baseWeights = map[string]float64{
	"metadata":   0.20,
	"heuristics": 0.30,
	"ai_model":   0.50,
}

func Compute(signals []entity.Signal) AggregationResult {
	available := make([]entity.Signal, 0, len(signals))
	var weightedSum float64
	var totalWeight float64

	for _, signal := range signals {
		if signal.Score < 0 {
			continue
		}

		weight := effectiveWeight(signal)
		if weight == 0 {
			continue
		}

		available = append(available, signal)
		weightedSum += float64(signal.Score) * weight
		totalWeight += weight
	}

	if totalWeight == 0 {
		return AggregationResult{
			FinalScore: -1,
			Confidence: entity.ConfidenceUnknown,
			Verdict:    VerdictUncertain,
			Signals:    signals,
		}
	}

	finalScore := weightedSum / totalWeight

	return AggregationResult{
		FinalScore: finalScore,
		Confidence: globalConfidence(available),
		Verdict:    verdict(finalScore),
		Signals:    signals,
	}
}

func AggregateMediaResults(results []AggregationResult) AggregationResult {
	if len(results) == 0 {
		return AggregationResult{
			FinalScore: -1,
			Confidence: entity.ConfidenceUnknown,
			Verdict:    VerdictUncertain,
		}
	}

	scores := make([]float64, 0, len(results))
	levels := make([]entity.ConfidenceLevel, 0, len(results)+1)
	var sum float64
	for _, result := range results {
		scores = append(scores, result.FinalScore)
		levels = append(levels, result.Confidence)
		sum += result.FinalScore
	}
	levels = append(levels, agreementConfidence(scores))

	finalScore := sum / float64(len(results))

	return AggregationResult{
		FinalScore: finalScore,
		Confidence: minConfidenceLevels(levels...),
		Verdict:    verdict(finalScore),
	}
}

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

func verdict(score float64) string {
	switch {
	case score < 30:
		return VerdictLikelyReal
	case score <= 70:
		return VerdictUncertain
	default:
		return VerdictLikelyAI
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
