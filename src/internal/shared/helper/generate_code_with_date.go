package helper

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
)

func GenerateCodeWithDate(tx *gorm.DB, model any, prefix string, createdAt time.Time) (string, error) {
	prefix = strings.ToUpper(strings.TrimSpace(prefix))
	tanggal := createdAt.Format("2006-01-02")
	tanggalCode := createdAt.Format("20060102")

	var lastCode sql.NullString
	err := tx.Model(model).
		Select("MAX(code)").
		Where("code LIKE ?", fmt.Sprintf("%s-%s%%", prefix, tanggalCode)).
		Where("DATE(created_at) = ?", tanggal).
		Scan(&lastCode).Error
	if err != nil {
		return "", err
	}

	lastNumber := 0
	if lastCode.Valid {
		// ambil 3 digit terakhir
		n, _ := strconv.Atoi(lastCode.String[len(lastCode.String)-3:])
		lastNumber = n
	}

	code := fmt.Sprintf("%s-%s%03d", prefix, tanggalCode, lastNumber+1)
	return code, nil
}
