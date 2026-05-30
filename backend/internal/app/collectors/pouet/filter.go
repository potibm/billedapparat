package pouet

import (
	"strings"

	"github.com/potibm/billedapparat/internal/app/contracts"
)

func filterByKeywords(slides []contracts.IngestSlideRequest, keywords []string) []contracts.IngestSlideRequest {
	var result []contracts.IngestSlideRequest

	for _, slide := range slides {
		if containsKeyword(slide.Body, keywords) {
			result = append(result, slide)
		}
	}

	return result
}

func containsKeyword(text string, keywords []string) bool {
	for _, keyword := range keywords {
		if containsIgnoreCase(text, keyword) {
			return true
		}
	}

	return false
}

func containsIgnoreCase(text, substring string) bool {
	return strings.Contains(strings.ToLower(text), substring)
}
