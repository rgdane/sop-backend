package helper

import (
	"strings"

	"gorm.io/datatypes"
)

func ExtractTextFromTiptap(content datatypes.JSONMap) string {
	text := ""

	if contentArray, ok := content["content"].([]interface{}); ok {
		for _, item := range contentArray {
			if itemMap, ok := item.(map[string]interface{}); ok {
				if innerContent, exists := itemMap["content"]; exists {
					if innerArray, ok := innerContent.([]interface{}); ok {
						for _, inner := range innerArray {
							if innerMap, ok := inner.(map[string]interface{}); ok {
								if t, ok := innerMap["text"].(string); ok {
									text += t + " "
								}
							}
						}
					}
				}
			}
		}
	}

	return strings.TrimSpace(text)
}
