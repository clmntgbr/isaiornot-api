package metadata

import (
	"encoding/binary"
	"strings"
)

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
