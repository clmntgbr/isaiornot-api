package heuristics

import (
	"io"

	"go-api/internal/domain/port"
)

type Analyzer struct{}

func NewAnalyzer() *Analyzer {
	return &Analyzer{}
}

func (a *Analyzer) Analyze(r io.Reader) (port.HeuristicsAnalysis, error) {
	data, err := io.ReadAll(io.LimitReader(r, maxScanBytes))
	if err != nil {
		return port.HeuristicsAnalysis{}, err
	}

	img, format, err := DecodeImage(data)
	if err != nil {
		return port.HeuristicsAnalysis{}, err
	}

	result := HeuristicsResult{
		NoiseScore:       AnalyzeNoise(img),
		CompressionScore: AnalyzeCompression(img, format, data),
		FrequencyScore:   AnalyzeFrequency(img),
		HistogramScore:   AnalyzeHistogram(img),
		Width:            img.Width,
		Height:           img.Height,
		Format:           format,
	}

	return result.ToAnalysis(), nil
}

var _ port.HeuristicsAnalyzer = (*Analyzer)(nil)
