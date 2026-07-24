package heuristics

import (
	"math"
	"math/cmplx"
)

const fftSize = 128

func AnalyzeFrequency(img *GrayscaleImage) float64 {
	sample := img.Downsample(fftSize)
	width := nextPowerOfTwo(min(sample.Width, fftSize))
	height := nextPowerOfTwo(min(sample.Height, fftSize))
	if width < 16 || height < 16 {
		return 0
	}

	surface := make([]float64, width*height)
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			surface[y*width+x] = sample.Pixel(
				x*sample.Width/width,
				y*sample.Height/height,
			)
		}
	}

	spectrum := fft2D(surface, width, height)
	return scoreFrequencySpectrum(spectrum, width, height)
}

func scoreFrequencySpectrum(spectrum []complex128, width, height int) float64 {
	centerX := width / 2
	centerY := height / 2

	var totalEnergy float64
	var highEnergy float64
	var peakEnergy float64

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			magnitude := cmplx.Abs(spectrum[y*width+x])
			totalEnergy += magnitude

			dx := math.Abs(float64(x - centerX))
			dy := math.Abs(float64(y - centerY))
			distance := math.Hypot(dx, dy)
			maxDistance := math.Hypot(float64(centerX), float64(centerY))

			if distance/maxDistance > 0.55 {
				highEnergy += magnitude
				if magnitude > peakEnergy {
					peakEnergy = magnitude
				}
			}
		}
	}

	if totalEnergy == 0 {
		return 0
	}

	highRatio := highEnergy / totalEnergy
	peakRatio := peakEnergy / (totalEnergy / float64(width*height))

	switch {
	case highRatio > 0.42 && peakRatio > 8:
		return 62
	case highRatio > 0.34 && peakRatio > 5:
		return 48
	case highRatio > 0.28:
		return 35
	default:
		return 18
	}
}

func fft2D(surface []float64, width, height int) []complex128 {
	rowTransformed := make([]complex128, width*height)
	for y := 0; y < height; y++ {
		row := make([]complex128, width)
		for x := 0; x < width; x++ {
			row[x] = complex(surface[y*width+x], 0)
		}
		row = fft1D(row)
		copy(rowTransformed[y*width:(y+1)*width], row)
	}

	result := make([]complex128, width*height)
	for x := 0; x < width; x++ {
		column := make([]complex128, height)
		for y := 0; y < height; y++ {
			column[y] = rowTransformed[y*width+x]
		}
		column = fft1D(column)
		for y := 0; y < height; y++ {
			result[y*width+x] = column[y]
		}
	}

	return result
}

func fft1D(input []complex128) []complex128 {
	n := len(input)
	if n <= 1 {
		return append([]complex128(nil), input...)
	}

	if n&(n-1) != 0 {
		next := nextPowerOfTwo(n)
		padded := make([]complex128, next)
		copy(padded, input)
		return fft1D(padded)[:n]
	}

	even := make([]complex128, n/2)
	odd := make([]complex128, n/2)
	for i := 0; i < n/2; i++ {
		even[i] = input[2*i]
		odd[i] = input[2*i+1]
	}

	even = fft1D(even)
	odd = fft1D(odd)

	output := make([]complex128, n)
	for k := 0; k < n/2; k++ {
		angle := -2 * math.Pi * float64(k) / float64(n)
		twiddle := cmplx.Exp(complex(0, angle))
		t := twiddle * odd[k]
		output[k] = even[k] + t
		output[k+n/2] = even[k] - t
	}

	return output
}

func nextPowerOfTwo(value int) int {
	if value <= 1 {
		return 1
	}

	power := 1
	for power < value {
		power <<= 1
	}

	return power
}

func min(a, b int) int {
	if a < b {
		return a
	}

	return b
}
