package enum

type AnalysisStatus string

const (
	AnalysisStatusProcessing AnalysisStatus = "processing"
	AnalysisStatusUploaded   AnalysisStatus = "uploaded"
	AnalysisStatusAnalyzed   AnalysisStatus = "analyzed"
	AnalysisStatusFailed     AnalysisStatus = "failed"
)
