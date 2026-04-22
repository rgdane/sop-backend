package helper

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