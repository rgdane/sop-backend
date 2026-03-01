package helper

import "reflect"

// RemoveUnchangedMap akan menghapus key dari payload jika nilainya sama dengan data lama.
// Dengan begitu GORM tidak akan mengupdate kolom yang tidak berubah.
func RemoveUnchangedFields(payload map[string]interface{}, oldData map[string]interface{}) map[string]interface{} {
	for k, v := range payload {
		if oldVal, ok := oldData[k]; ok {
			// kalau sama → hapus
			if reflect.DeepEqual(oldVal, v) {
				delete(payload, k)
			}
		}
	}
	return payload
}
