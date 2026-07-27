package enum

type ScanStatus string

const (
	ScanStatusUploaded   ScanStatus = "uploaded"
	ScanStatusProcessing ScanStatus = "processing"
	ScanStatusAnalyzed   ScanStatus = "analyzed"
	ScanStatusFailed     ScanStatus = "failed"
)
