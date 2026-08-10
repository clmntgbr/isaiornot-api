package aggregate

import (
	"math"

	domainscan "go-api/internal/domain/scan"
	domainsignal "go-api/internal/domain/signal"
)

const (
	VerdictLikelyReal = "likely_real"
	VerdictUncertain  = "uncertain"
	VerdictLikelyAI   = "likely_ai"
)

type InputSignal struct {
	Name       string
	Score      int
	Confidence domainsignal.ConfidenceLevel
}

type Result struct {
	FinalScore float64
	Confidence domainscan.ConfidenceLevel
	Verdict    string
	HasAIModel bool
}

var baseWeights = map[string]float64{
	string(domainsignal.NameMetadata):    0.20,
	string(domainsignal.NameHeuristics):  0.30,
	string(domainsignal.NameSightEngine): 0.50,
}

func Compute(signals []InputSignal) Result {
	var weightedSum float64
	var totalWeight float64
	available := make([]InputSignal, 0, len(signals))
	hasAIModel := false

	for _, signal := range signals {
		if signal.Score < 0 {
			continue
		}
		weight := effectiveWeight(signal)
		if weight == 0 {
			continue
		}
		if signal.Name == string(domainsignal.NameSightEngine) {
			hasAIModel = true
		}
		available = append(available, signal)
		weightedSum += float64(signal.Score) * weight
		totalWeight += weight
	}

	if totalWeight == 0 {
		return Result{
			FinalScore: -1,
			Confidence: domainscan.ConfidenceUncertain,
			Verdict:    VerdictUncertain,
			HasAIModel: hasAIModel,
		}
	}

	finalScore := weightedSum / totalWeight
	return Result{
		FinalScore: finalScore,
		Confidence: toScanConfidence(globalConfidence(available)),
		Verdict:    verdict(finalScore, hasAIModel, available),
		HasAIModel: hasAIModel,
	}
}

func AggregateMediaResults(results []Result) Result {
	if len(results) == 0 {
		return Result{
			FinalScore: -1,
			Confidence: domainscan.ConfidenceUncertain,
			Verdict:    VerdictUncertain,
		}
	}

	var sum float64
	levels := make([]domainscan.ConfidenceLevel, 0, len(results))
	hasAIModel := true
	for _, result := range results {
		sum += result.FinalScore
		levels = append(levels, result.Confidence)
		if !result.HasAIModel {
			hasAIModel = false
		}
	}

	finalScore := sum / float64(len(results))
	return Result{
		FinalScore: finalScore,
		Confidence: minScanConfidence(levels...),
		Verdict:    verdictFromScore(finalScore, hasAIModel),
		HasAIModel: hasAIModel,
	}
}

func effectiveWeight(signal InputSignal) float64 {
	base, ok := baseWeights[signal.Name]
	if !ok {
		return 0
	}
	weight := base * confidenceMultiplier(signal.Confidence)
	if signal.Name == string(domainsignal.NameSightEngine) && signal.Confidence == domainsignal.ConfidenceHigh {
		weight *= 1.5
	}
	return weight
}

func confidenceMultiplier(confidence domainsignal.ConfidenceLevel) float64 {
	switch confidence {
	case domainsignal.ConfidenceHigh:
		return 1.0
	case domainsignal.ConfidenceMedium:
		return 0.75
	case domainsignal.ConfidenceLow:
		return 0.5
	default:
		return 0.25
	}
}

func globalConfidence(available []InputSignal) domainsignal.ConfidenceLevel {
	count := len(available)
	if count == 0 {
		return domainsignal.ConfidenceUnknown
	}

	scores := make([]float64, 0, count)
	hasHighAI := false
	for _, signal := range available {
		scores = append(scores, float64(signal.Score))
		if signal.Name == string(domainsignal.NameSightEngine) && signal.Confidence == domainsignal.ConfidenceHigh {
			hasHighAI = true
		}
	}
	_, std := meanAndStd(scores)

	switch {
	case count >= 3 && std <= 20:
		return domainsignal.ConfidenceHigh
	case count >= 3 && std <= 35:
		return domainsignal.ConfidenceMedium
	case hasHighAI && count >= 2 && std <= 40:
		return domainsignal.ConfidenceMedium
	case count >= 2:
		return domainsignal.ConfidenceLow
	default:
		return domainsignal.ConfidenceLow
	}
}

func verdict(score float64, hasAIModel bool, available []InputSignal) string {
	if !hasAIModel {
		if !hasReliableNonAIEvidence(available) {
			return VerdictUncertain
		}
		if score < 50 {
			return VerdictLikelyReal
		}
		return VerdictLikelyAI
	}
	return verdictFromScore(score, true)
}

func verdictFromScore(score float64, hasAIModel bool) string {
	if !hasAIModel {
		return VerdictUncertain
	}
	switch {
	case score < 30:
		return VerdictLikelyReal
	case score <= 70:
		return VerdictUncertain
	default:
		return VerdictLikelyAI
	}
}

func hasReliableNonAIEvidence(available []InputSignal) bool {
	for _, signal := range available {
		if signal.Name == string(domainsignal.NameSightEngine) {
			continue
		}
		if signal.Confidence == domainsignal.ConfidenceMedium || signal.Confidence == domainsignal.ConfidenceHigh {
			return true
		}
		if signal.Name == string(domainsignal.NameMetadata) && signal.Score >= 0 && signal.Score < 30 {
			return true
		}
	}
	return false
}

func toScanConfidence(level domainsignal.ConfidenceLevel) domainscan.ConfidenceLevel {
	switch level {
	case domainsignal.ConfidenceHigh:
		return domainscan.ConfidenceHigh
	case domainsignal.ConfidenceMedium:
		return domainscan.ConfidenceMedium
	case domainsignal.ConfidenceLow:
		return domainscan.ConfidenceLow
	default:
		return domainscan.ConfidenceUncertain
	}
}

func minScanConfidence(levels ...domainscan.ConfidenceLevel) domainscan.ConfidenceLevel {
	if len(levels) == 0 {
		return domainscan.ConfidenceUncertain
	}
	rank := func(level domainscan.ConfidenceLevel) int {
		switch level {
		case domainscan.ConfidenceHigh:
			return 3
		case domainscan.ConfidenceMedium:
			return 2
		case domainscan.ConfidenceLow:
			return 1
		default:
			return 0
		}
	}
	min := levels[0]
	for _, level := range levels[1:] {
		if rank(level) < rank(min) {
			min = level
		}
	}
	return min
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
