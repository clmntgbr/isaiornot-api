package heuristics

import (
	"math"
)

func AnalyzeHistogram(img *GrayscaleImage) float64 {
	if len(img.Data) == 0 {
		return 0
	}

	lumaEntropy := channelEntropy(img.Data)
	saturationScore := analyzeSaturationUniformity(img)
	sharpnessScore := analyzeSharpness(img)

	return clampScore(0.45*lumaEntropy + 0.30*saturationScore + 0.25*sharpnessScore)
}

func channelEntropy(values []float64) float64 {
	bins := make([]int, 32)
	for _, value := range values {
		bin := int(value / 8)
		if bin >= len(bins) {
			bin = len(bins) - 1
		}
		if bin < 0 {
			bin = 0
		}
		bins[bin]++
	}

	var entropy float64
	total := float64(len(values))
	for _, count := range bins {
		if count == 0 {
			continue
		}
		p := float64(count) / total
		entropy -= p * math.Log2(p)
	}

	maxEntropy := math.Log2(float64(len(bins)))
	if maxEntropy == 0 {
		return 50
	}

	normalized := entropy / maxEntropy

	switch {
	case normalized < 0.72:
		return 62
	case normalized < 0.82:
		return 48
	case normalized < 0.90:
		return 35
	default:
		return 22
	}
}

func analyzeSaturationUniformity(img *GrayscaleImage) float64 {
	var localRanges []float64
	window := 12

	for y := 0; y <= img.Height-window; y += window {
		for x := 0; x <= img.Width-window; x += window {
			minValue := math.MaxFloat64
			maxValue := -math.MaxFloat64
			for dy := 0; dy < window; dy++ {
				for dx := 0; dx < window; dx++ {
					value := img.Pixel(x+dx, y+dy)
					if value < minValue {
						minValue = value
					}
					if value > maxValue {
						maxValue = value
					}
				}
			}
			localRanges = append(localRanges, maxValue-minValue)
		}
	}

	_, std := meanAndStd(localRanges)
	if std < 12 {
		return 58
	}
	if std < 20 {
		return 42
	}

	return 24
}

func analyzeSharpness(img *GrayscaleImage) float64 {
	if img.Width < 3 || img.Height < 3 {
		return 0
	}

	var laplacianValues []float64
	for y := 1; y < img.Height-1; y++ {
		for x := 1; x < img.Width-1; x++ {
			center := img.Pixel(x, y)
			laplacian := -4*center +
				img.Pixel(x-1, y) +
				img.Pixel(x+1, y) +
				img.Pixel(x, y-1) +
				img.Pixel(x, y+1)
			laplacianValues = append(laplacianValues, math.Abs(laplacian))
		}
	}

	mean, std := meanAndStd(laplacianValues)
	if mean == 0 {
		return 40
	}

	ratio := std / mean

	switch {
	case ratio < 0.45:
		return 60
	case ratio < 0.70:
		return 45
	case ratio < 1.00:
		return 32
	default:
		return 20
	}
}
