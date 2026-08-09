package video

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

const MaxFrames = 10

type FrameExtractor struct{}

func NewFrameExtractor() *FrameExtractor {
	return &FrameExtractor{}
}

func (e *FrameExtractor) ExtractFrames(videoData []byte, maxFrames int) ([][]byte, error) {
	if maxFrames <= 0 {
		maxFrames = MaxFrames
	}

	tmpDir, err := os.MkdirTemp("", "scan-video-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	inputPath := filepath.Join(tmpDir, "input")
	if err := os.WriteFile(inputPath, videoData, 0o600); err != nil {
		return nil, fmt.Errorf("failed to write temp video: %w", err)
	}

	duration, err := probeDuration(inputPath)
	if err != nil {
		return nil, err
	}
	if duration <= 0 {
		duration = 0.1
	}

	timestamps := sampleTimestamps(duration, maxFrames)
	frames := make([][]byte, 0, len(timestamps))

	for i, ts := range timestamps {
		framePath := filepath.Join(tmpDir, fmt.Sprintf("frame-%02d.jpg", i))
		if err := extractFrameAt(inputPath, framePath, ts); err != nil {
			return nil, fmt.Errorf("failed to extract frame at %.2fs: %w", ts, err)
		}

		frameBytes, err := os.ReadFile(framePath)
		if err != nil {
			return nil, fmt.Errorf("failed to read frame %d: %w", i, err)
		}
		frames = append(frames, frameBytes)
	}

	return frames, nil
}

func sampleTimestamps(duration float64, maxFrames int) []float64 {
	if maxFrames < 1 {
		maxFrames = 1
	}

	timestamps := make([]float64, 0, maxFrames)
	for i := 0; i < maxFrames; i++ {
		ts := duration * (float64(i) + 0.5) / float64(maxFrames)
		if ts >= duration {
			ts = duration * 0.99
		}
		if ts < 0 {
			ts = 0
		}
		timestamps = append(timestamps, ts)
	}

	return timestamps
}

func probeDuration(path string) (float64, error) {
	cmd := exec.Command(
		"ffprobe",
		"-v", "error",
		"-show_entries", "format=duration",
		"-of", "default=noprint_wrappers=1:nokey=1",
		path,
	)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return 0, fmt.Errorf("ffprobe failed: %w (%s)", err, strings.TrimSpace(stderr.String()))
	}

	raw := strings.TrimSpace(stdout.String())
	duration, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid duration %q: %w", raw, err)
	}

	return duration, nil
}

func extractFrameAt(inputPath, outputPath string, timestamp float64) error {
	cmd := exec.Command(
		"ffmpeg",
		"-y",
		"-ss", fmt.Sprintf("%.3f", timestamp),
		"-i", inputPath,
		"-frames:v", "1",
		"-q:v", "2",
		outputPath,
	)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("ffmpeg failed: %w (%s)", err, strings.TrimSpace(stderr.String()))
	}

	return nil
}
