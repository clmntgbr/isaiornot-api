package metadata

import (
	"io"

	"go-api/domain/entity"
)

const maxScanBytes = MaxScanBytes

type ScanResult struct {
	Signal    entity.Signal     `json:"signal"`
	Extracted ExtractedMetadata `json:"extracted"`
}

type Analyzer struct{}

func NewAnalyzer() *Analyzer {
	return &Analyzer{}
}

func (a *Analyzer) Analyze(r io.Reader) (ScanResult, error) {
	data, err := io.ReadAll(io.LimitReader(r, maxScanBytes))
	if err != nil {
		return ScanResult{}, err
	}

	extracted := Extract(data)
	signal := Score(extracted)

	return ScanResult{
		Signal:    signal,
		Extracted: extracted,
	}, nil
}
