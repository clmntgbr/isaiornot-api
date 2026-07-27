package port

type FrameExtractor interface {
	ExtractFrames(videoData []byte, maxFrames int) ([][]byte, error)
}
