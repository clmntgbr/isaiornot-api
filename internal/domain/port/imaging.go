package port

import (
	"context"
	"io"
)

type ImageThumbnailer interface {
	GenerateJPEG(ctx context.Context, src io.Reader, maxWidth int) ([]byte, error)
}
