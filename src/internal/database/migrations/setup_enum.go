package migrations

import (
	"fmt"
	"strings"

	"gorm.io/gorm"
)

type EnumDefinition struct {
	Name   string
	Values []string
}

var enumList = []EnumDefinition{
	{Name: "draft_enum", Values: []string{"draft", "published"}},
	{Name: "shift_type_enum", Values: []string{"technical", "customer_service"}},
}

func SetupEnumUp(db *gorm.DB) error {
	for _, enum := range enumList {
		values := "'" + strings.Join(enum.Values, "','") + "'"

		query := fmt.Sprintf(
			`CREATE TYPE IF NOT EXISTS %s AS ENUM (%s);`,
			enum.Name,
			values,
		)

		if err := db.Exec(query).Error; err != nil {
			return err
		}
	}
	return nil
}

func SetupEnumDown(db *gorm.DB) error {
	for _, enum := range enumList {
		query := fmt.Sprintf(
			`DROP TYPE IF EXISTS %s CASCADE;`,
			enum.Name,
		)

		if err := db.Exec(query).Error; err != nil {
			return err
		}
	}
	return nil
}
