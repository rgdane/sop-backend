package helper

import (
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

func Neo4jFormatter(data []neo4j.Record) any {
	if len(data) == 0 {
		return nil
	}
	formattedData := make([]any, len(data))
	for i, record := range data {
		if len(record.Values) == 1 {
			formattedData[i] = record.Values[0]
		} else {
			formattedData[i] = record.Values
		}
	}
	return formattedData
}

// func Neo4jFormatter(data []neo4j.Record) any {
// 	if len(data) == 0 {
// 		return nil
// 	}
// 	formattedData := make([]any, len(data))
// 	for i, record := range data {
// thReturn		if len(record.Values) == 1 {
// 			formattedData[i] = record.Values[0]
// 		} else {
// 			formattedData[i] = record.Values
// 		}
// 	}
// 	return formattedData
// }
