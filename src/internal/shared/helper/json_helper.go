package helper

import "encoding/json"

// ToJSONString mengubah value map/slice menjadi string JSON
func ToJSONString(v interface{}) string {
	// Kalau v sudah string, jangan marshal lagi
	if s, ok := v.(string); ok {
		return s
	}

	b, err := json.Marshal(v)
	if err != nil {
		return ""
	}
	return string(b)
}
