package metadata

import (
	"regexp"
	"strings"
)

var wordBoundaryPatterns = map[string]bool{
	"flux":    true,
	"gemini":  true,
	"imagen":  true,
	"openai":  true,
	"copilot": true,
	"dalle":   true,
	"firefly": true,
}

var boundaryRegexCache = map[string]*regexp.Regexp{}

func signatureMatches(haystack string, signature SoftwareSignature) bool {
	normalizedHaystack := strings.ToLower(haystack)
	pattern := strings.ToLower(signature.Pattern)

	if wordBoundaryPatterns[pattern] {
		return boundaryPattern(pattern).MatchString(normalizedHaystack)
	}

	return strings.Contains(normalizedHaystack, pattern)
}

func boundaryPattern(pattern string) *regexp.Regexp {
	if cached, ok := boundaryRegexCache[pattern]; ok {
		return cached
	}

	re := regexp.MustCompile(`(?i)\b` + regexp.QuoteMeta(pattern) + `\b`)
	boundaryRegexCache[pattern] = re
	return re
}

func matchesNeutralSoftware(values ...string) (bool, string) {
	joined := strings.ToLower(strings.Join(values, " "))
	for _, neutral := range NeutralSoftware {
		if strings.Contains(joined, neutral) {
			return true, neutral
		}
	}

	return false, ""
}
