package scan

import (
	"go-api/domain/aggregate"
	"go-api/domain/entity"
	"go-api/domain/enum"
	"go-api/domain/paginate"
)

type ListQuery struct {
	paginate.PaginateQuery
	Status     string `query:"status"`
	Confidence string `query:"confidence"`
	Verdict    string `query:"verdict"`
}

func (q *ListQuery) Normalize() {
	q.PaginateQuery.Normalize()

	switch enum.ScanStatus(q.Status) {
	case enum.ScanStatusUploaded, enum.ScanStatusProcessing, enum.ScanStatusCompleted, enum.ScanStatusFailed:
	default:
		q.Status = ""
	}

	switch entity.ConfidenceLevel(q.Confidence) {
	case entity.ConfidenceHigh, entity.ConfidenceMedium, entity.ConfidenceLow, entity.ConfidenceUnknown:
	default:
		q.Confidence = ""
	}

	switch q.Verdict {
	case aggregate.VerdictLikelyReal, aggregate.VerdictUncertain, aggregate.VerdictLikelyAI:
	default:
		q.Verdict = ""
	}
}
