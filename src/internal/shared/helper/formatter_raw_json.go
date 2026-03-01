package helper

import "encoding/json"

func RawJSONFormatter(data interface{}) *json.RawMessage {
	dataBytes, err := json.Marshal(data)
	if err != nil {
		return nil
	}

	dataJSON := json.RawMessage(dataBytes)

	return &dataJSON
}
