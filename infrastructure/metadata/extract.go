package metadata

import (
	"bytes"
	"encoding/binary"
	"regexp"
	"strings"

	"github.com/rwcarlsen/goexif/exif"
)

const MaxScanBytes = 512 * 1024

const maxMetadataTextBytes = 64 * 1024

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

func extractEXIF(data []byte, result *ExtractedMetadata) {
	x, err := exif.Decode(bytes.NewReader(data))
	if err != nil {
		return
	}

	tags := map[exif.FieldName]string{
		exif.Software: "",
		exif.Make:     "",
		exif.Model:    "",
		exif.Artist:   "",
	}

	for field := range tags {
		tag, err := x.Get(field)
		if err != nil {
			continue
		}

		value, err := tag.StringVal()
		if err != nil {
			continue
		}

		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}

		result.RawTags[string(field)] = value

		switch field {
		case exif.Software:
			result.Software = value
		case exif.Make:
			result.Make = value
		case exif.Model:
			result.Model = value
		case exif.Artist:
			result.Artist = value
		}
	}
}

func extractXMPBlocks(data []byte) []string {
	var blocks []string

	if isJPEG(data) {
		blocks = append(blocks, extractJPEGXMP(data)...)
	}

	blocks = append(blocks, extractEmbeddedXMP(data)...)

	seen := make(map[string]struct{})
	unique := make([]string, 0, len(blocks))
	for _, block := range blocks {
		normalized := strings.TrimSpace(block)
		if normalized == "" {
			continue
		}
		if _, ok := seen[normalized]; ok {
			continue
		}
		seen[normalized] = struct{}{}
		unique = append(unique, normalized)
	}

	return unique
}

func isJPEG(data []byte) bool {
	return len(data) >= 2 && data[0] == 0xFF && data[1] == 0xD8
}

func extractJPEGXMP(data []byte) []string {
	var blocks []string
	offset := 2

	for offset+4 < len(data) {
		if data[offset] != 0xFF {
			break
		}

		marker := data[offset+1]
		if marker == 0xD9 {
			break
		}

		length := int(binary.BigEndian.Uint16(data[offset+2 : offset+4]))
		if length < 2 || offset+2+length > len(data) {
			break
		}

		segment := data[offset+4 : offset+2+length]

		switch marker {
		case 0xE1:
			if xmp := xmpFromAPP1(segment); xmp != "" {
				blocks = append(blocks, xmp)
			}
		case 0xEB:
			if xmp := xmpFromAPP11(segment); xmp != "" {
				blocks = append(blocks, xmp)
			}
		}

		offset += 2 + length
	}

	return blocks
}

func xmpFromAPP1(segment []byte) string {
	prefixes := []string{
		"http://ns.adobe.com/xap/1.0/\x00",
		"http://ns.adobe.com/xap/1.0/",
		"XMP\x00",
	}

	for _, prefix := range prefixes {
		if strings.HasPrefix(string(segment), prefix) {
			return string(segment[len(prefix):])
		}
	}

	return ""
}

func xmpFromAPP11(segment []byte) string {
	content := string(segment)
	if strings.Contains(content, "xmp") || strings.Contains(content, "XMP") {
		return content
	}

	return ""
}

func extractEmbeddedXMP(data []byte) []string {
	content := string(data)
	markers := []string{"<x:xmpmeta", "<rdf:RDF", "<?xpacket"}

	var blocks []string
	for _, marker := range markers {
		start := strings.Index(content, marker)
		if start == -1 {
			continue
		}

		end := strings.Index(content[start:], "</x:xmpmeta>")
		if end != -1 {
			blocks = append(blocks, content[start:start+end+len("</x:xmpmeta>")])
			continue
		}

		end = strings.Index(content[start:], "</rdf:RDF>")
		if end != -1 {
			blocks = append(blocks, content[start:start+end+len("</rdf:RDF>")])
		}
	}

	return blocks
}

var (
	xmpTagPattern           = regexp.MustCompile(`(?i)<(?:[^:>]+:)?([A-Za-z0-9_]+)[^>]*>([^<]{1,512})</`)
	promptTagPattern        = regexp.MustCompile(`(?i)(?:prompt|parameters|description)\s*[=:]\s*"?([^"\n]{4,512})"?`)
	seedTagPattern          = regexp.MustCompile(`(?i)seed\s*[=:]\s*"?([0-9]{1,20})"?`)
	negativePromptPattern   = regexp.MustCompile(`(?i)negative prompt:\s*([^\n]+)`)
	sdParametersLinePattern = regexp.MustCompile(`(?i)(steps|sampler|cfg scale|model hash|model|version|clip skip)\s*:`)
)

func parseMetadataBlock(keyword, block, source string, result *ExtractedMetadata) {
	normalized := strings.ToLower(block)
	summary := summarizeMetadata(block)
	result.XMPSnippets = appendUnique(result.XMPSnippets, summarizeMetadataSource(source, keyword, summary))

	if keyword != "" {
		result.RawTags[source+":"+strings.ToLower(keyword)] = summary
	}

	parseGenerationChunk(keyword, block, source, result)

	if result.CreatorTool == "" {
		result.CreatorTool = firstXMPValue(block, []string{"creatortool", "xmp:creatortool"})
	}

	if result.Software == "" && result.CreatorTool != "" {
		result.Software = result.CreatorTool
	}

	if result.Prompt == "" {
		result.Prompt = firstXMPValue(block, []string{"prompt", "parameters", "dc:description"})
	}

	if result.Seed == "" {
		result.Seed = firstXMPValue(block, []string{"seed"})
	}

	for _, indicator := range PromptIndicators {
		if strings.Contains(normalized, indicator) {
			result.RawTags["meta:"+indicator] = "present"
		}
	}

	if match := promptTagPattern.FindStringSubmatch(block); len(match) > 1 && result.Prompt == "" {
		result.Prompt = strings.TrimSpace(match[1])
	}

	if match := seedTagPattern.FindStringSubmatch(block); len(match) > 1 && result.Seed == "" {
		result.Seed = strings.TrimSpace(match[1])
	}

	matches := xmpTagPattern.FindAllStringSubmatch(block, -1)
	for _, match := range matches {
		if len(match) < 3 {
			continue
		}

		key := strings.ToLower(match[1])
		value := strings.TrimSpace(match[2])
		if value == "" {
			continue
		}

		tagKey := "xmp:" + key
		if _, exists := result.RawTags[tagKey]; !exists {
			result.RawTags[tagKey] = value
		}
	}
}

func parseGenerationChunk(keyword, block, source string, result *ExtractedMetadata) {
	lowerKeyword := strings.ToLower(keyword)
	switch lowerKeyword {
	case "parameters", "prompt", "workflow", "description":
	default:
		if !sdParametersLinePattern.MatchString(block) {
			return
		}
	}

	if result.Prompt == "" {
		switch lowerKeyword {
		case "parameters":
			result.Prompt = extractA1111Prompt(block)
		case "prompt":
			result.Prompt = strings.TrimSpace(block)
		}
	}

	if lowerKeyword == "workflow" {
		result.RawTags[source+":workflow"] = "present"
	}

	if match := negativePromptPattern.FindStringSubmatch(block); len(match) > 1 {
		result.RawTags[source+":negative_prompt"] = strings.TrimSpace(match[1])
	}
}

func extractA1111Prompt(block string) string {
	lines := strings.Split(block, "\n")
	if len(lines) == 0 {
		return ""
	}

	firstLine := strings.TrimSpace(lines[0])
	if firstLine == "" {
		return ""
	}

	if strings.HasPrefix(strings.ToLower(firstLine), "negative prompt:") {
		return ""
	}

	return firstLine
}

func summarizeMetadataSource(source, keyword, summary string) string {
	if keyword == "" {
		return source + ": " + summary
	}

	return source + ":" + keyword + ": " + summary
}

func firstXMPValue(block string, keys []string) string {
	lowerBlock := strings.ToLower(block)
	for _, key := range keys {
		key = strings.ToLower(key)
		tagPattern := regexp.MustCompile(`(?i)<[^>]*:?` + regexp.QuoteMeta(strings.TrimPrefix(key, "xmp:")) + `[^>]*>([^<]+)</`)
		if match := tagPattern.FindStringSubmatch(lowerBlock); len(match) > 1 {
			return strings.TrimSpace(match[1])
		}

		attrPattern := regexp.MustCompile(`(?i)` + regexp.QuoteMeta(key) + `="([^"]+)"`)
		if match := attrPattern.FindStringSubmatch(block); len(match) > 1 {
			return strings.TrimSpace(match[1])
		}
	}

	return ""
}

func summarizeMetadata(block string) string {
	compact := strings.Join(strings.Fields(block), " ")
	if len(compact) > 240 {
		return compact[:240] + "..."
	}

	return compact
}

func scanBinaryMarkers(data []byte, result *ExtractedMetadata) {
	lower := strings.ToLower(string(data))

	if strings.Contains(lower, "c2pa") || strings.Contains(lower, "jumbf") || strings.Contains(lower, "contentcredentials") {
		result.C2PA = true
	}

	for _, indicator := range C2PAAIIndicators {
		if strings.Contains(lower, indicator) {
			result.C2PA = true
			result.C2PADetails = appendUnique(result.C2PADetails, indicator)
		}
	}

	for _, generator := range C2PAGenerators {
		if strings.Contains(lower, generator) && result.C2PA {
			result.C2PADetails = appendUnique(result.C2PADetails, "generator:"+generator)
		}
	}
}

func appendUnique(values []string, value string) []string {
	for _, existing := range values {
		if existing == value {
			return values
		}
	}

	return append(values, value)
}
