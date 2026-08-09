package insight

import (
	"time"

	"github.com/google/uuid"
)

type Insight struct {
	ID          uuid.UUID
	Noise       float64
	Compression float64
	Frequency   float64
	Histogram   float64
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func NewInsight(noise, compression, frequency, histogram float64) *Insight {
	now := time.Now().UTC()
	return &Insight{
		ID:          uuid.New(),
		Noise:       noise,
		Compression: compression,
		Frequency:   frequency,
		Histogram:   histogram,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

func (i *Insight) UpdateScores(noise, compression, frequency, histogram float64) {
	i.Noise = noise
	i.Compression = compression
	i.Frequency = frequency
	i.Histogram = histogram
	i.UpdatedAt = time.Now().UTC()
}
