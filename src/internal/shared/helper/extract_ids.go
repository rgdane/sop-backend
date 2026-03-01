package helper

// ExtractIDsInt64 mengekstrak ID dari slice struct yang punya field ID ke slice int64
func ExtractIDsInt64[T any](items []T, getID func(T) int64) *[]int64 {
	var ids []int64
	for _, item := range items {
		ids = append(ids, getID(item))
	}
	return &ids
}
