package metadata

import "strings"

type ExtractedMetadata struct {
	Software    string            `json:"software,omitempty"`
	Make        string            `json:"make,omitempty"`
	Model       string            `json:"model,omitempty"`
	Artist      string            `json:"artist,omitempty"`
	CreatorTool string            `json:"creator_tool,omitempty"`
	Prompt      string            `json:"prompt,omitempty"`
	Seed        string            `json:"seed,omitempty"`
	C2PA        bool              `json:"c2pa"`
	C2PADetails []string          `json:"c2pa_details,omitempty"`
	XMPSnippets []string          `json:"xmp_snippets,omitempty"`
	RawTags     map[string]string `json:"raw_tags,omitempty"`
}

func (m ExtractedMetadata) HasAnyMetadata() bool {
	return m.Software != "" ||
		m.Make != "" ||
		m.Model != "" ||
		m.Artist != "" ||
		m.CreatorTool != "" ||
		m.Prompt != "" ||
		m.Seed != "" ||
		m.C2PA ||
		len(m.XMPSnippets) > 0 ||
		len(m.RawTags) > 0
}

func (m ExtractedMetadata) HasCameraTags() bool {
	return m.Make != "" || m.Model != ""
}

const MaxScanBytes = 512 * 1024

const maxMetadataTextBytes = 64 * 1024

func Extract(data []byte) ExtractedMetadata {
	result := ExtractedMetadata{
		RawTags: make(map[string]string),
	}

	extractEXIF(data, &result)

	for _, block := range extractXMPBlocks(data) {
		parseMetadataBlock("", block, "xmp", &result)
	}

	for _, chunk := range extractPNGTextChunks(data) {
		parseMetadataBlock(chunk.keyword, chunk.text, chunk.chunkType, &result)
	}

	scanBinaryMarkers(data, &result)

	return result
}

func summarizeMetadataSource(source, keyword, summary string) string {
	if keyword == "" {
		return source + ": " + summary
	}

	return source + ":" + keyword + ": " + summary
}

func summarizeMetadata(block string) string {
	compact := strings.Join(strings.Fields(block), " ")
	if len(compact) > 240 {
		return compact[:240] + "..."
	}

	return compact
}

func appendUnique(values []string, value string) []string {
	for _, existing := range values {
		if existing == value {
			return values
		}
	}

	return append(values, value)
}
