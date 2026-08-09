package metadata

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"fmt"
	"io"
)

const pngSignature = "\x89PNG\r\n\x1a\n"

type pngTextChunk struct {
	chunkType string
	keyword   string
	text      string
}

func isPNG(data []byte) bool {
	return len(data) >= len(pngSignature) && string(data[:len(pngSignature)]) == pngSignature
}

func extractPNGTextChunks(data []byte) []pngTextChunk {
	if !isPNG(data) {
		return nil
	}

	var chunks []pngTextChunk
	offset := len(pngSignature)

	for offset+12 <= len(data) {
		length := int(binary.BigEndian.Uint32(data[offset : offset+4]))
		chunkType := string(data[offset+4 : offset+8])

		dataStart := offset + 8
		dataEnd := dataStart + length
		if dataEnd+4 > len(data) {
			break
		}

		chunkData := data[dataStart:dataEnd]

		switch chunkType {
		case "tEXt":
			if chunk, err := parsePNGTextChunk(chunkType, chunkData); err == nil {
				chunks = append(chunks, chunk)
			}
		case "zTXt":
			if chunk, err := parsePNGZTextChunk(chunkType, chunkData); err == nil {
				chunks = append(chunks, chunk)
			}
		case "iTXt":
			if chunk, err := parsePNGiTextChunk(chunkType, chunkData); err == nil {
				chunks = append(chunks, chunk)
			}
		}

		if chunkType == "IEND" {
			break
		}

		offset = dataEnd + 4
	}

	return chunks
}

func parsePNGTextChunk(chunkType string, data []byte) (pngTextChunk, error) {
	keyword, text, ok := splitPNGKeywordValue(data)
	if !ok {
		return pngTextChunk{}, fmt.Errorf("invalid tEXt chunk")
	}

	return pngTextChunk{
		chunkType: chunkType,
		keyword:   keyword,
		text:      text,
	}, nil
}

func parsePNGZTextChunk(chunkType string, data []byte) (pngTextChunk, error) {
	nullIdx := bytes.IndexByte(data, 0)
	if nullIdx <= 0 || nullIdx+2 > len(data) {
		return pngTextChunk{}, fmt.Errorf("invalid zTXt chunk")
	}

	keyword := string(data[:nullIdx])
	compressed := data[nullIdx+2:]

	text, err := decompressPNGText(compressed)
	if err != nil {
		return pngTextChunk{}, err
	}

	return pngTextChunk{
		chunkType: chunkType,
		keyword:   keyword,
		text:      text,
	}, nil
}

func parsePNGiTextChunk(chunkType string, data []byte) (pngTextChunk, error) {
	nullIdx := bytes.IndexByte(data, 0)
	if nullIdx <= 0 || nullIdx+2 > len(data) {
		return pngTextChunk{}, fmt.Errorf("invalid iTXt chunk")
	}

	keyword := string(data[:nullIdx])
	compressed := data[nullIdx+1] == 1
	cursor := nullIdx + 2

	for cursor < len(data) && data[cursor] != 0 {
		cursor++
	}
	if cursor >= len(data) {
		return pngTextChunk{}, fmt.Errorf("invalid iTXt language tag")
	}
	cursor++

	for cursor < len(data) && data[cursor] != 0 {
		cursor++
	}
	if cursor >= len(data) {
		return pngTextChunk{}, fmt.Errorf("invalid iTXt translated keyword")
	}
	cursor++

	payload := data[cursor:]
	var text string
	var err error

	if compressed {
		text, err = decompressPNGText(payload)
		if err != nil {
			return pngTextChunk{}, err
		}
	} else {
		text = string(payload)
	}

	return pngTextChunk{
		chunkType: chunkType,
		keyword:   keyword,
		text:      text,
	}, nil
}

func splitPNGKeywordValue(data []byte) (string, string, bool) {
	nullIdx := bytes.IndexByte(data, 0)
	if nullIdx <= 0 || nullIdx >= len(data)-1 {
		return "", "", false
	}

	keyword := string(data[:nullIdx])
	if keyword == "" {
		return "", "", false
	}

	return keyword, string(data[nullIdx+1:]), true
}

func decompressPNGText(data []byte) (string, error) {
	reader, err := zlib.NewReader(bytes.NewReader(data))
	if err != nil {
		return "", err
	}
	defer reader.Close()

	decompressed, err := io.ReadAll(io.LimitReader(reader, maxMetadataTextBytes))
	if err != nil {
		return "", err
	}

	return string(decompressed), nil
}
