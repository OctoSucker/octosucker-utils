package utils

import (
	"encoding/json"
	"regexp"
	"strings"
)

// ExtractJSONFromMarkdown 从 markdown 代码块中提取 JSON
func ExtractJSONFromMarkdown(text string) string {
	// 匹配 ```json ... ``` 或 ``` ... ```
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

// ExtractJSONObject 从文本中提取第一个有效的 JSON 对象
func ExtractJSONObject(text string) string {
	text = strings.TrimSpace(text)

	// 尝试直接解析整个文本
	if IsValidJSON(text) {
		return text
	}

	// 查找第一个 { 和最后一个 }
	startIdx := strings.Index(text, "{")
	if startIdx == -1 {
		return ""
	}

	// 从 startIdx 开始，找到匹配的 }
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

	// 如果找不到匹配的 }，尝试找到最后一个 }
	endIdx := strings.LastIndex(text, "}")
	if endIdx > startIdx {
		jsonStr := text[startIdx : endIdx+1]
		if IsValidJSON(jsonStr) {
			return jsonStr
		}
	}

	return ""
}

// IsValidJSON 检查字符串是否是有效的 JSON
func IsValidJSON(s string) bool {
	var v interface{}
	return json.Unmarshal([]byte(s), &v) == nil
}
