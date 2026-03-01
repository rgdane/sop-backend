package models

import (
	"time"

	"gorm.io/gorm"
)

type DatabaseNode struct {
	ID        int64          `gorm:"primarykey;autoIncrement:false;type:bigint;default:nextval('database_node_seq'::regclass)" json:"id"`
	Name      string         `gorm:"column:name;not null" json:"name"`
	TableRef  string         `gorm:"column:table_ref" json:"table_ref"`
	GraphRef  string         `gorm:"column:graph_ref" json:"graph_ref"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at"`
}

func (DatabaseNode) TableName() string {
	return "database_nodes"
}
