package media

type Status string

const (
	StatusPending    Status = "pending"
	StatusUploaded   Status = "uploaded"
	StatusProcessing Status = "processing"
	StatusCompleted  Status = "completed"
	StatusFailed     Status = "failed"
)
