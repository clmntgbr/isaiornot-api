package dto

type ObjectCreatedEvent struct {
	Records []ObjectCreatedRecord `json:"Records"`
}

type ObjectCreatedRecord struct {
	EventName string   `json:"eventName"`
	S3        S3Entity `json:"s3"`
}

type S3Entity struct {
	Bucket S3Bucket `json:"bucket"`
	Object S3Object `json:"object"`
}

type S3Bucket struct {
	Name string `json:"name"`
}

type S3Object struct {
	Key         string `json:"key"`
	Size        int64  `json:"size"`
	ContentType string `json:"contentType"`
}
