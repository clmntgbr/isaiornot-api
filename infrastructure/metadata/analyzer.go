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

func (a *Analyzer) Analyze(r io.Reader) (entity.Signal, error) {
	data, err := io.ReadAll(io.LimitReader(r, maxScanBytes))
	if err != nil {
		return entity.Signal{}, err
	}

	extracted := Extract(data)
	return Score(extracted), nil
}
