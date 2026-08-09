package heuristics

import (
	"bytes"
	"image"
	"image/color"
	"image/jpeg"
	"math"
)

const jpegBlockSize = 8

func AnalyzeCompression(img *GrayscaleImage, format string, original []byte) float64 {
	blockScore := analyzeBlockBoundaries(img)

	if format != "jpeg" {
		return clampScore(blockScore * 0.75)
	}

	elaScore := analyzeELA(img, original)
	return clampScore(0.55*blockScore + 0.45*elaScore)
}

func analyzeBlockBoundaries(img *GrayscaleImage) float64 {
	if img.Width < jpegBlockSize*2 || img.Height < jpegBlockSize*2 {
		return 0
	}

	var boundaryDiffs []float64
	var internalDiffs []float64

	for y := jpegBlockSize; y < img.Height-jpegBlockSize; y += jpegBlockSize {
		for x := 1; x < img.Width-1; x++ {
			boundaryDiffs = append(boundaryDiffs, math.Abs(img.Pixel(x, y)-img.Pixel(x, y-1)))
		}
	}

	for y := 1; y < img.Height-1; y++ {
		for x := jpegBlockSize; x < img.Width-jpegBlockSize; x += jpegBlockSize {
			internalDiffs = append(internalDiffs, math.Abs(img.Pixel(x, y)-img.Pixel(x-1, y)))
		}
	}

	_, boundaryStd := meanAndStd(boundaryDiffs)
	_, internalStd := meanAndStd(internalDiffs)
	if internalStd == 0 {
		return 40
	}

	ratio := boundaryStd / internalStd

	switch {
	case ratio < 0.65:
		return 70
	case ratio < 0.85:
		return 55
	case ratio < 1.10:
		return 35
	default:
		return 20
	}
}

func analyzeELA(img *GrayscaleImage, original []byte) float64 {
	rgba := grayscaleToRGBA(img)

	var buffer bytes.Buffer
	if err := jpeg.Encode(&buffer, rgba, &jpeg.Options{Quality: 95}); err != nil {
		return 40
	}

	recompressed, err := jpeg.Decode(bytes.NewReader(buffer.Bytes()))
	if err != nil {
		return 40
	}

	_ = original

	var residuals []float64
	bounds := rgba.Bounds()
	reBounds := recompressed.Bounds()

	for y := 0; y < bounds.Dy(); y++ {
		for x := 0; x < bounds.Dx(); x++ {
			origGray := img.Pixel(x, y)
			r, g, b, _ := recompressed.At(reBounds.Min.X+x, reBounds.Min.Y+y).RGBA()
			reGray := 0.299*float64(r>>8) + 0.587*float64(g>>8) + 0.114*float64(b>>8)
			residuals = append(residuals, math.Abs(origGray-reGray))
		}
	}

	mean, std := meanAndStd(residuals)
	if mean == 0 {
		return 35
	}

	uniformity := std / mean

	switch {
	case uniformity < 0.35:
		return 68
	case uniformity < 0.55:
		return 52
	case uniformity < 0.80:
		return 38
	default:
		return 22
	}
}

func grayscaleToRGBA(img *GrayscaleImage) *image.RGBA {
	rgba := image.NewRGBA(image.Rect(0, 0, img.Width, img.Height))
	for y := 0; y < img.Height; y++ {
		for x := 0; x < img.Width; x++ {
			value := uint8(math.Round(img.Pixel(x, y)))
			rgba.Set(x, y, color.RGBA{R: value, G: value, B: value, A: 255})
		}
	}

	return rgba
}
