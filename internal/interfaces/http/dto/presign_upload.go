package dto

type PresignUploadRequest struct {
	Filename    string `json:"filename"`
	ContentType string `json:"contentType"`
}
