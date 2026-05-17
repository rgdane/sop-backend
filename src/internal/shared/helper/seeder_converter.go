package helper

import (
	"math/rand"
	"time"
	"fmt"
)

// Helper function untuk mengubah string menjadi *string (dibutuhkan oleh Description)
func StrPtr(s string) *string {
	return &s
}

func Int64Ptr(i int64) *int64 {
	return &i
}

func BoolPtr(b bool) *bool {
	return &b
}

// Helper untuk generate warna Hex acak untuk Title Color
func GenerateRandomHexColor() string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return fmt.Sprintf("#%06x", r.Intn(0xFFFFFF))
}

// Helper untuk generate array ID acak (tanpa duplikat)
func GenerateRandomIDs(min, max, count int) []int64 {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	uniqueIDs := make(map[int]bool)
	var result []int64

	for len(result) < count {
		val := r.Intn(max-min+1) + min
		if !uniqueIDs[val] {
			uniqueIDs[val] = true
			result = append(result, int64(val))
		}
	}
	return result
}