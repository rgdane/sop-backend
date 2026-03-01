package helper

import "reflect"

// convert struct ke map, misal model.backlog ke map[string]interface{}
func StructToMap(data interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	val := reflect.ValueOf(data)
	typ := reflect.TypeOf(data)

	if val.Kind() == reflect.Ptr {
		val = val.Elem()
		typ = typ.Elem()
	}

	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		tag := typ.Field(i).Tag.Get("json")
		if tag == "" || tag == "-" {
			continue
		}

		if field.CanInterface() {
			value := field.Interface()

			// kalau field pointer -> ambil nilai aslinya
			if field.Kind() == reflect.Ptr && !field.IsNil() {
				value = field.Elem().Interface()
			}

			result[tag] = value
		}
	}

	return result
}
