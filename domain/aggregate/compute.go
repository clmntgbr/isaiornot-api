package aggregate

import "go-api/domain/entity"

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
