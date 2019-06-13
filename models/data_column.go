package models

import (
	"github.com/jinzhu/gorm"
	"supermentor/db"
)

type DataColumn struct {
	gorm.Model
	AbilityID        uint
	Key              string
	Name             string
	Desc             string
	Unit             string
	Order            int
	Type             string  // 类型
	LowestValue      int     `json:"-"` // 最低值
	GamaValue        int     `json:"-"` // gama值
	Weighing         float64 `json:"-"` // 勤奋值权重
	AbilityWeighing  float64 `json:"-"` // 能力度权重
	DiligentWeighing float64 `json:"-"` // 勤奋度权重
	RecommendFormula string  `json:"-"` // 推荐公式
}

var cachedDataColumns = make([]DataColumn, 0)

func (DataColumn) TableName() string {
	return "data_column"
}

func GetAllDataColumnToArray() []DataColumn {
	if len(cachedDataColumns) == 0 {
		db.Db.Order("`order` ASC").Find(&cachedDataColumns)
	}
	return cachedDataColumns
}

// 查询所有的数据字段
func GetAllDataColumn() map[uint]DataColumn {
	if len(cachedDataColumns) == 0 {
		db.Db.Order("`order` ASC").Find(&cachedDataColumns)
	}
	idKeys := make(map[uint]DataColumn)
	// 组合成DataId => DataColumns的map
	for i := 0; i < len(cachedDataColumns); i++ {
		c := cachedDataColumns[i]
		idKeys[c.ID] = c
	}
	// 查询成功
	return idKeys
}
