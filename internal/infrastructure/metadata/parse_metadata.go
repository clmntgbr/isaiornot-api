package metadata

import (
	"regexp"
	"strings"
)

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
