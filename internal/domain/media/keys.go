package media

import (
	"fmt"
	"mime"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

type PresignUploadInput struct {
	Filename    string
	ContentType string
}

var imageExtensions = map[string]struct{}{
	"jpg":  {},
	"jpeg": {},
	"png":  {},
	"webp": {},
}

var videoExtensions = map[string]struct{}{
	"mp4":  {},
	"mov":  {},
	"avi":  {},
	"mkv":  {},
	"m4v":  {},
	"mpeg": {},
	"mpg":  {},
	"wmv":  {},
	"asf":  {},
	"flv":  {},
	"webm": {},
	"ogg":  {},
	"ogv":  {},
	"mka":  {},
}

func fileExtension(filename string) string {
	return strings.ToLower(strings.TrimPrefix(filepath.Ext(filename), "."))
}

func ValidatePresignUploadInput(input PresignUploadInput) error {
	ext := fileExtension(input.Filename)
	if ext == "" {
		return fmt.Errorf("filename must have a supported extension")
	}

	if !IsImageFilename(input.Filename) && !IsVideoFilename(input.Filename) {
		return fmt.Errorf("unsupported file type: .%s", ext)
	}

	return nil
}

func IsImageFilename(filename string) bool {
	_, ok := imageExtensions[fileExtension(filename)]
	return ok
}

func IsVideoFilename(filename string) bool {
	_, ok := videoExtensions[fileExtension(filename)]
	return ok
}

func ContentTypeFromKey(key string, fallback string) string {
	if fallback != "" {
		return fallback
	}

	if contentType := mime.TypeByExtension(strings.ToLower(filepath.Ext(key))); contentType != "" {
		return contentType
	}

	return "application/octet-stream"
}

func IsImageContentType(contentType string) bool {
	return strings.HasPrefix(contentType, "image/")
}

func IsVideoContentType(contentType string) bool {
	return strings.HasPrefix(contentType, "video/")
}

const maxVideoFrames = 10

func MaxVideoFrames() int {
	return maxVideoFrames
}

func NewFrameFileKey() string {
	return "frames/" + uuid.NewString() + ".jpg"
}

func IsFrameObjectKey(decodedKey string) bool {
	return strings.Contains(decodedKey, "/frames/")
}

func NewFileKey(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	return uuid.NewString() + ext
}

func NewObjectKey(userID, scanID uuid.UUID, fileKey string) string {
	return userID.String() + "/" + scanID.String() + "/" + fileKey
}

func NewObjectKeyFromFilename(userID, scanID uuid.UUID, filename string) string {
	return NewObjectKey(userID, scanID, NewFileKey(filename))
}

func NewThumbnailFileKey(mediaID uuid.UUID) string {
	return "thumbnails/" + mediaID.String() + ".jpg"
}

func NewThumbnailObjectKey(userID, scanID, mediaID uuid.UUID) string {
	return NewObjectKey(userID, scanID, NewThumbnailFileKey(mediaID))
}

func DecodeObjectKey(key string) (string, error) {
	decoded, err := url.QueryUnescape(key)
	if err != nil {
		return "", fmt.Errorf("invalid media key: %w", err)
	}

	return decoded, nil
}

func UserIDFromKey(encodedKey string) (uuid.UUID, error) {
	key, err := DecodeObjectKey(encodedKey)
	if err != nil {
		return uuid.Nil, err
	}

	userID, _, ok := strings.Cut(key, "/")
	if !ok {
		return uuid.Nil, fmt.Errorf("invalid media key: %q", key)
	}

	return uuid.Parse(userID)
}

func ScanIDFromKey(encodedKey string) (uuid.UUID, error) {
	key, err := DecodeObjectKey(encodedKey)
	if err != nil {
		return uuid.Nil, err
	}

	parts := strings.SplitN(key, "/", 3)
	if len(parts) < 3 {
		return uuid.Nil, fmt.Errorf("invalid media key: %q", key)
	}

	return uuid.Parse(parts[1])
}

func FileKeyFromObjectKey(encodedKey string) (string, error) {
	key, err := DecodeObjectKey(encodedKey)
	if err != nil {
		return "", err
	}

	parts := strings.SplitN(key, "/", 3)
	if len(parts) < 3 || parts[2] == "" {
		return "", fmt.Errorf("invalid media key: %q", key)
	}

	return parts[2], nil
}
