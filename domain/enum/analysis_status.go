package enum

type AnalysisStatus string

const (
	AnalysisStatusUploaded   AnalysisStatus = "uploaded"
	AnalysisStatusProcessing AnalysisStatus = "processing"
	AnalysisStatusAnalyzed   AnalysisStatus = "analyzed"
	AnalysisStatusFailed     AnalysisStatus = "failed"
)
