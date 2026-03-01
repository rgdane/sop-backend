package helper

import (
	"fmt"
	"strings"
)

func GenerateCode(name string, id int64) string {
	prefix := strings.ToLower(name)
	if len(prefix) > 3 {
		prefix = prefix[:3]
	}

	code := fmt.Sprintf("%s-%03d", prefix, id)

	return code
}
