package models

import (
	"github.com/jinzhu/gorm"
	"supermentor/db"
)

type Ability struct {
	gorm.Model
	Title string `gorm:"type:varchar(100)"`
}

func (Ability) TableName() string {
	return "ability"
}

var cachedAbility = make([]Ability, 0)

// 查询所有的数据字段
func GetAllAbility() map[uint]Ability {
	if len(cachedAbility) == 0 {
		db.Db.Order("`id` ASC").Find(&cachedAbility)
	}
	idKeys := make(map[uint]Ability)
	// 组合成 DataId => Ability 的map
	for i := 0; i < len(cachedAbility); i++ {
		c := cachedAbility[i]
		idKeys[c.ID] = c
	}
	// 查询成功
	return idKeys
}