package utils

import (
	"encoding/json"
	"regexp"
	"strings"
)

func ExtractJSONFromMarkdown(text string) string {
	patterns := []*regexp.Regexp{
		regexp.MustCompile("(?s)```json\\s*(.*?)\\s*```"),
		regexp.MustCompile("(?s)```\\s*(.*?)\\s*```"),
	}

	for _, pattern := range patterns {
		matches := pattern.FindStringSubmatch(text)
		if len(matches) > 1 {
			jsonStr := strings.TrimSpace(matches[1])
			if IsValidJSON(jsonStr) {
				return jsonStr
			}
		}
	}

	return ""
}

func ExtractJSONObject(text string) string {
	text = strings.TrimSpace(text)

	if IsValidJSON(text) {
		return text
	}

	startIdx := strings.Index(text, "{")
	if startIdx == -1 {
		return ""
	}

	depth := 0
	for i := startIdx; i < len(text); i++ {
		if text[i] == '{' {
			depth++
		} else if text[i] == '}' {
			depth--
			if depth == 0 {
				jsonStr := text[startIdx : i+1]
				if IsValidJSON(jsonStr) {
					return jsonStr
				}
			}
		}
	}

	endIdx := strings.LastIndex(text, "}")
	if endIdx > startIdx {
		jsonStr := text[startIdx : endIdx+1]
		if IsValidJSON(jsonStr) {
			return jsonStr
		}
	}

	return ""
}

func IsValidJSON(s string) bool {
	var v interface{}
	return json.Unmarshal([]byte(s), &v) == nil
}
