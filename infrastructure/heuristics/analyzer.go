package heuristics

import (
	"io"
)

type Analyzer struct{}

func NewAnalyzer() *Analyzer {
	return &Analyzer{}
}

func (a *Analyzer) Analyze(r io.Reader) (ScanResult, error) {
	data, err := io.ReadAll(io.LimitReader(r, maxScanBytes))
	if err != nil {
		return ScanResult{}, err
	}

	img, format, err := DecodeImage(data)
	if err != nil {
		return ScanResult{}, err
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

	return ScanResult{
		Signal:     heuristics.ToSignal(),
		Heuristics: heuristics,
	}, nil
}
