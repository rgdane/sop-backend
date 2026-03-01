package helper

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
)

func ParseQueryInt64(c *fiber.Ctx, key string) (int64, error) {
	param := c.Query(key)
	if param == "" {
		return 0, nil
	}
	parsed, err := strconv.ParseInt(param, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("Invalid %s", key)
	}
	return parsed, nil
}

func ParseQueryArrayInt(c *fiber.Ctx, key string) ([]int64, error) {
	var result []int64

	// Ambil semua nilai `key` dari query
	queryArgs := c.Request().URI().QueryArgs()
	values := queryArgs.PeekMulti(key) // [][]byte

	for _, v := range values {
		val, err := strconv.ParseInt(string(v), 10, 64)
		if err != nil {
			return nil, err
		}
		result = append(result, val)
	}

	return result, nil
}

func ParseQueryInt64Array(c *fiber.Ctx, key string) ([]int64, error) {
	val := c.Query(key)
	if val == "" {
		return nil, nil
	}

	var result []int64

	// Coba parse sebagai JSON array
	if strings.HasPrefix(val, "[") && strings.HasSuffix(val, "]") {
		if err := json.Unmarshal([]byte(val), &result); err == nil {
			return result, nil
		}
	}

	// Kalau bukan JSON array, coba split manual (comma atau semicolon)
	sep := ","
	if strings.Contains(val, ";") {
		sep = ";"
	}

	parts := strings.Split(val, sep)
	for _, p := range parts {
		p = strings.TrimSpace(strings.Trim(p, "[]"))
		if p == "" {
			continue
		}
		if id, err := strconv.ParseInt(p, 10, 64); err == nil {
			result = append(result, id)
		}
	}

	return result, nil
}

// For multiple string filter
func ParseQueryStringArray(c *fiber.Ctx, key string) ([]string, error) {
	val := c.Query(key)

	var list []string
	if val != "" {
		list = strings.Split(val, ",")
	} else {
		return nil, fmt.Errorf("tidak ada filter status")
	}

	return list, nil
}
