package media

type Status string

const (
	StatusProcessing Status = "processing"
	StatusUploaded   Status = "uploaded"
	StatusCompleted  Status = "completed"
	StatusFailed     Status = "failed"
)
