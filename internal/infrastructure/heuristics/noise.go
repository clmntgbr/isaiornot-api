package heuristics

const noisePatchSize = 16

func AnalyzeNoise(img *GrayscaleImage) float64 {
	if img.Width < noisePatchSize || img.Height < noisePatchSize {
		return 0
	}

	var patchVariances []float64
	for y := 0; y <= img.Height-noisePatchSize; y += noisePatchSize {
		for x := 0; x <= img.Width-noisePatchSize; x += noisePatchSize {
			variance := patchVariance(img.Data, img.Width, img.Height, noisePatchSize, x, y)
			patchVariances = append(patchVariances, variance)
		}
	}

	if len(patchVariances) == 0 {
		return 0
	}

	mean, std := meanAndStd(patchVariances)
	if mean <= 0 {
		return 50
	}

	coefficientOfVariation := std / mean

	switch {
	case coefficientOfVariation < 0.25:
		return 72
	case coefficientOfVariation < 0.40:
		return 58
	case coefficientOfVariation < 0.60:
		return 42
	default:
		return 25
	}
}
