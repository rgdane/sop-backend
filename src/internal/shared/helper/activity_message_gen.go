package helper

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"
)

type LogParams struct {
	Action      string
	TableRef    string
	TableRefId  *int64
	TableRefIds *[]int64
	Ctx         *fiber.Ctx
	OldData     *json.RawMessage
	NewData     *json.RawMessage
}

func GenerateActivityMessage(action, user string, newdata, olddata *json.RawMessage) string {
	switch action {
	case "DELETE":
		if newdata == nil { // bulk delete
			result := multipleDeleteDiff(olddata, newdata)
			return fmt.Sprintf("<p>%s deleted multiple:<br>%s</p>", user, result)
		}
		name := getNameFromJSON(olddata)
		return fmt.Sprintf("<p>%s deleted %s</p>", user, name)
	case "CREATE":
		name := getNameFromJSON(newdata)
		return fmt.Sprintf("<p>%s created %s</p>", user, name)
	case "UPDATE":
		if olddata != nil && newdata != nil { // single update
			result := buildUpdateDiff(olddata, newdata)
			name := getNameFromJSON(olddata)
			return fmt.Sprintf("<p>%s updated %s:<br>%s</p>", user, name, result)
		}
		// bulk update
		result := multipleUpdateDiff(olddata, newdata)
		return fmt.Sprintf("<p>%s updated multiple:<br>%s</p>", user, result)
	default:
		return "<p>Action was not found</p>"
	}
}

func getNameFromJSON(data *json.RawMessage) string {
	if data == nil {
		return "unknown"
	}

	// Try to unmarshal as single object first
	var m map[string]interface{}
	if err := json.Unmarshal(*data, &m); err == nil {
		return extractNameFromMap(m)
	}

	// If that fails, try to unmarshal as array
	var arr []map[string]interface{}
	if err := json.Unmarshal(*data, &arr); err == nil {
		if len(arr) > 0 {
			return extractNameFromMap(arr[0])
		}
	}

	// Try as array of any interface (for mixed types)
	var anyArr []interface{}
	if err := json.Unmarshal(*data, &anyArr); err == nil {
		if len(anyArr) > 0 {
			if obj, ok := anyArr[0].(map[string]interface{}); ok {
				return extractNameFromMap(obj)
			}
		}
	}

	return "unknown"
}

func extractNameFromMap(m map[string]interface{}) string {
	// Try multiple possible name fields in order of preference
	nameFields := []string{"name", "title", "summary", "description", "label"}

	for _, field := range nameFields {
		if val, exists := m[field]; exists && val != nil {
			if str := fmt.Sprint(val); str != "" && str != "<nil>" {
				return str
			}
		}
	}

	// If no name field found, try to use ID as fallback
	if id, exists := m["id"]; exists && id != nil {
		return fmt.Sprintf("Item #%v", id)
	}

	return "unknown"
}

func buildUpdateDiff(oldjson, newjson *json.RawMessage) string {
	if oldjson == nil || newjson == nil {
		return ""
	}

	var oldMap, newMap map[string]interface{}
	if err := json.Unmarshal(*oldjson, &oldMap); err != nil {
		return ""
	}
	if err := json.Unmarshal(*newjson, &newMap); err != nil {
		return ""
	}

	var b strings.Builder
	for key, newVal := range newMap {
		if oldVal, ok := oldMap[key]; ok {
			if fmt.Sprint(oldVal) != fmt.Sprint(newVal) && key != "updated_at" {
				b.WriteString(fmt.Sprintf(
					`<div>%s: <span style="opacity:50%%"><s>%v</s></span> %v</div>`,
					key, oldVal, newVal,
				))
			}
		}
	}
	return b.String()
}

func multipleUpdateDiff(oldjson, newjson *json.RawMessage) string {
	if oldjson == nil || newjson == nil {
		return ""
	}

	var oldList []map[string]interface{}
	var newObj map[string]interface{}

	if err := json.Unmarshal(*oldjson, &oldList); err != nil {
		return ""
	}
	if err := json.Unmarshal(*newjson, &newObj); err != nil {
		return ""
	}

	var b strings.Builder
	for _, oldItem := range oldList {
		name := extractNameFromMap(oldItem)
		b.WriteString(fmt.Sprintf(`<div><strong>Item: %s</strong></div>`, name))

		for key, newVal := range newObj {
			if key == "updated_at" || newVal == nil {
				continue
			}
			oldVal := oldItem[key]
			if fmt.Sprint(oldVal) != fmt.Sprint(newVal) {
				b.WriteString(fmt.Sprintf(
					`<div>%s: <span style="opacity:50%%"><s>%v</s></span> %v</div>`,
					key, oldVal, newVal,
				))
			}
		}
		b.WriteString("<br>")
	}

	return b.String()
}

func multipleDeleteDiff(oldjson, newjson *json.RawMessage) string {
	if oldjson == nil {
		return ""
	}

	var oldList []map[string]interface{}
	if err := json.Unmarshal(*oldjson, &oldList); err != nil {
		return ""
	}

	var b strings.Builder
	for _, oldItem := range oldList {
		name := extractNameFromMap(oldItem)
		b.WriteString(fmt.Sprintf(`<div><strong>Deleted Item: %s</strong></div>`, name))
	}
	return b.String()
}

// Keep the old function for backward compatibility
func getDisplayName(data map[string]interface{}) string {
	return extractNameFromMap(data)
}
