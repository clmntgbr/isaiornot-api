package scan

type Status string

const (
	StatusUploaded   Status = "uploaded"
	StatusProcessing Status = "processing"
	StatusCompleted  Status = "completed"
	StatusFailed     Status = "failed"
)

type ConfidenceLevel string

const (
	ConfidenceUncertain ConfidenceLevel = "uncertain"
	ConfidenceLow       ConfidenceLevel = "low"
	ConfidenceMedium    ConfidenceLevel = "medium"
	ConfidenceHigh      ConfidenceLevel = "high"
)
