package extractor

import (
	"encoding/json"
	"regexp"
	"strings"
)

// urlFields lists JSON-LD / Schema.org keys whose string values are always IRIs.
var urlFields = map[string]bool{
	"@id": true,
	"url": true,
}

func ParseJSONLD(URL string, htmlContent string) ([]map[string]any, []error) {
	items, errs := extractJSONLD(htmlContent)
	if URL != "" {
		for _, item := range items {
			resolveJSONLDURLs(item, URL)
		}
	}
	return items, errs
}

func extractJSONLD(htmlContent string) ([]map[string]any, []error) {
	re := regexp.MustCompile(`(?s)<script[^>]+type[ \t\n]*=[ \t\n]*["']application/ld\+json["'][^>]*>(.*?)</script>`)

	matches := re.FindAllStringSubmatch(htmlContent, -1)

	var errors []error
	var jsonLDs []map[string]any
	for _, match := range matches {
		if len(match) > 1 {
			jsonLD := strings.TrimSpace(match[1])
			if jsonLD != "" {
				if jsonLD[0] == '[' {
					var jsonData []map[string]any
					if err := json.Unmarshal([]byte(jsonLD), &jsonData); err != nil {
						errors = append(errors, err)
					} else {
						jsonLDs = append(jsonLDs, jsonData...)
					}
				} else if jsonLD[0] == '{' {
					var jsonData map[string]any
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

func resolveJSONLDURLs(obj map[string]any, baseURL string) {
	for key, val := range obj {
		switch v := val.(type) {
		case string:
			if urlFields[key] {
				obj[key] = resolveURL(baseURL, v)
			}
		case map[string]any:
			resolveJSONLDURLs(v, baseURL)
		case []any:
			resolveJSONLDArray(v, baseURL)
		}
	}
}

func resolveJSONLDArray(arr []any, baseURL string) {
	for _, item := range arr {
		if obj, ok := item.(map[string]any); ok {
			resolveJSONLDURLs(obj, baseURL)
		}
	}
}
