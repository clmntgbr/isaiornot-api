package heuristics

import (
	"io"

	"go-api/domain/port"
)

type Analyzer struct{}

func NewAnalyzer() *Analyzer {
	return &Analyzer{}
}

func (a *Analyzer) Analyze(r io.Reader) (port.HeuristicsScanResult, error) {
	data, err := io.ReadAll(io.LimitReader(r, maxScanBytes))
	if err != nil {
		return port.HeuristicsScanResult{}, err
	}

	img, format, err := DecodeImage(data)
	if err != nil {
		return port.HeuristicsScanResult{}, err
	}

	heuristics := HeuristicsResult{
		NoiseScore:       AnalyzeNoise(img),
		CompressionScore: AnalyzeCompression(img, format, data),
		FrequencyScore:   AnalyzeFrequency(img),
		HistogramScore:   AnalyzeHistogram(img),
		Width:            img.Width,
		Height:           img.Height,
		Format:           format,
	}

	return port.HeuristicsScanResult{
		Signal:           heuristics.ToSignal(),
		NoiseScore:       heuristics.NoiseScore,
		CompressionScore: heuristics.CompressionScore,
		FrequencyScore:   heuristics.FrequencyScore,
		HistogramScore:   heuristics.HistogramScore,
	}, nil
}
