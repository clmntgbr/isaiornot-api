package metadata

import (
	"io"

	"go-api/internal/domain/port"
)

type Analyzer struct{}

func NewAnalyzer() *Analyzer {
	return &Analyzer{}
}

func (a *Analyzer) Analyze(r io.Reader) (port.MetadataAnalysis, error) {
	data, err := io.ReadAll(io.LimitReader(r, MaxScanBytes))
	if err != nil {
		return port.MetadataAnalysis{}, err
	}

	extracted := Extract(data)
	return Score(extracted), nil
}

var _ port.MetadataAnalyzer = (*Analyzer)(nil)
