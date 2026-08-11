package dto

type PresignUploadRequest struct {
	Filename    string `json:"filename" validate:"required,min=1,max=255"`
	ContentType string `json:"contentType" validate:"omitempty,max=127"`
}
