package heuristics

import (
	"bytes"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"math"

	_ "golang.org/x/image/webp"
)

const (
	maxScanBytes      = 8 * 1024 * 1024
	minReliablePixels = 128 * 128
)

type GrayscaleImage struct {
	Width  int
	Height int
	Data   []float64
}

func (g *GrayscaleImage) Pixel(x, y int) float64 {
	if x < 0 || y < 0 || x >= g.Width || y >= g.Height {
		return 0
	}

	return g.Data[y*g.Width+x]
}

func DecodeImage(data []byte) (*GrayscaleImage, string, error) {
	if len(data) > maxScanBytes {
		data = data[:maxScanBytes]
	}

	img, format, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, "", fmt.Errorf("failed to decode image: %w", err)
	}

	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()
	if width == 0 || height == 0 {
		return nil, format, fmt.Errorf("invalid image dimensions")
	}

	gray := &GrayscaleImage{
		Width:  width,
		Height: height,
		Data:   make([]float64, width*height),
	}

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			r, g, b, _ := img.At(bounds.Min.X+x, bounds.Min.Y+y).RGBA()
			luma := 0.299*float64(r>>8) + 0.587*float64(g>>8) + 0.114*float64(b>>8)
			gray.Data[y*width+x] = luma
		}
	}

	return gray, format, nil
}

func (g *GrayscaleImage) Downsample(maxSize int) *GrayscaleImage {
	if g.Width <= maxSize && g.Height <= maxSize {
		return g
	}

	scale := float64(maxSize) / float64(max(g.Width, g.Height))
	newWidth := max(1, int(math.Round(float64(g.Width)*scale)))
	newHeight := max(1, int(math.Round(float64(g.Height)*scale)))

	result := &GrayscaleImage{
		Width:  newWidth,
		Height: newHeight,
		Data:   make([]float64, newWidth*newHeight),
	}

	for y := 0; y < newHeight; y++ {
		srcY := int(float64(y) * float64(g.Height) / float64(newHeight))
		for x := 0; x < newWidth; x++ {
			srcX := int(float64(x) * float64(g.Width) / float64(newWidth))
			result.Data[y*newWidth+x] = g.Pixel(srcX, srcY)
		}
	}

	return result
}

func patchVariance(data []float64, width, height, patchSize, startX, startY int) float64 {
	var sum, sumSq float64
	count := 0

	for y := startY; y < startY+patchSize && y < height; y++ {
		for x := startX; x < startX+patchSize && x < width; x++ {
			value := data[y*width+x]
			sum += value
			sumSq += value * value
			count++
		}
	}

	if count == 0 {
		return 0
	}

	mean := sum / float64(count)
	return sumSq/float64(count) - mean*mean
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

func clampScore(value float64) float64 {
	if value < 0 {
		return 0
	}
	if value > 100 {
		return 100
	}

	return value
}

func max(a, b int) int {
	if a > b {
		return a
	}

	return b
}
