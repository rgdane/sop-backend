package helper

func GetIdentifierProps(labels []string, allProps map[string]any) map[string]any {
	identifiers := make(map[string]any)

	// Tentukan identifier berdasarkan label
	for _, label := range labels {
		switch label {
		case "Job":
			// Untuk Job: gunakan index + tableRef sebagai unique identifier
			// Atau bisa juga parent + index
			if index, ok := allProps["index"]; ok {
				identifiers["index"] = index
			}
			if tableRef, ok := allProps["tableRef"]; ok {
				identifiers["tableRef"] = tableRef
			}

		case "Column":
			// Untuk Column: gunakan index sebagai unique identifier
			// Karena sudah ada relationship dari parent (Table/Job)
			if index, ok := allProps["index"]; ok {
				identifiers["index"] = index
			}

		case "Table":
			// Untuk Table: gunakan name sebagai identifier
			if name, ok := allProps["name"]; ok {
				identifiers["name"] = name
			}

		case "Row":
			// Untuk Row: gunakan rowIndex + productId
			if rowIndex, ok := allProps["rowIndex"]; ok {
				identifiers["rowIndex"] = rowIndex
			}
			if productId, ok := allProps["productId"]; ok {
				identifiers["productId"] = productId
			}

		case "Document":
			// Untuk Document: gunakan id
			if id, ok := allProps["id"]; ok {
				identifiers["id"] = id
			}

		case "Comment":
			// Untuk Comment: gunakan commentId
			if commentId, ok := allProps["commentId"]; ok {
				identifiers["commentId"] = commentId
			}

		case "Review":
			// Untuk Review: gunakan reviewId
			if reviewId, ok := allProps["reviewId"]; ok {
				identifiers["reviewId"] = reviewId
			}

		default:
			// Default: gunakan name sebagai identifier jika ada
			if name, ok := allProps["name"]; ok {
				identifiers["name"] = name
			}
		}
	}

	// Jika tidak ada identifier yang ditemukan, gunakan semua props (fallback ke behavior lama)
	if len(identifiers) == 0 {
		return allProps
	}

	return identifiers
}
