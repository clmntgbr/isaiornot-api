package enum

type ScanStatus string

const (
	ScanStatusUploaded   ScanStatus = "uploaded"
	ScanStatusProcessing ScanStatus = "processing"
	ScanStatusCompleted  ScanStatus = "completed"
	ScanStatusFailed     ScanStatus = "failed"
)
