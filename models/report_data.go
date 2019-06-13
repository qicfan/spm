package models

import (
	"github.com/jinzhu/gorm"
)

type ReportData struct {
	gorm.Model
	UserID uint
	ReportID uint
	Value float64
	DataColumnID uint
	AbilityNum float64
	MaxUser uint
	Rank uint
	ReportDataColumn DataColumn `gorm:"ForeignKey:DataColumnID"` // 关联的数据项
}

func (ReportData) TableName() string {
	return "report_data"
}