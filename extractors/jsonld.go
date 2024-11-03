package extractor

import (
	"encoding/json"
	"regexp"
	"strings"
)

func JSONLD(URL string, htmlContent string) ([]map[string]interface{}, []error) {
	_ = URL
	items, errors := extractJSONLD(htmlContent)

	var results []map[string]interface{}
	if len(items) >= 0 {
		results = append(results, items...)
	}

	return results, errors
}

func extractJSONLD(htmlContent string) ([]map[string]interface{}, []error) {
	re := regexp.MustCompile(`(?s)<script[^>]+type=["']application/ld\+json["'][^>]*>(.*?)</script>`)

	matches := re.FindAllStringSubmatch(htmlContent, -1)

	var errors []error
	var jsonLDs []map[string]interface{}
	for _, match := range matches {
		if len(match) > 1 {
			jsonLD := strings.TrimSpace(match[1])
			if jsonLD != "" {
				if jsonLD[0] == '[' {
					var jsonData []map[string]interface{}
					if err := json.Unmarshal([]byte(jsonLD), &jsonData); err != nil {
						errors = append(errors, err)
					} else {
						jsonLDs = append(jsonLDs, jsonData...)
					}
				} else if jsonLD[0] == '{' {
					var jsonData map[string]interface{}
					if err := json.Unmarshal([]byte(jsonLD), &jsonData); err != nil {
						errors = append(errors, err)
					} else {
						jsonLDs = append(jsonLDs, jsonData)
					}
				}
			}
		}
	}

	return jsonLDs, errors
}
