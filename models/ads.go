package models

import (
	"github.com/jinzhu/gorm"
	"supermentor/db"
)

type Ads struct {
	gorm.Model
	Position  string `gorm:"type:varchar(100)"`
	Cover     string `gorm:"type:varchar(500)"`
	Action    int
	ActionUrl string `gorm:"type:varchar(500)"`
}

func (Ads) TableName() string {
	return "ads"
}

func GetAds(position string) []Ads {
	ads := make([]Ads, 0)
	db.Db.Where("position=?", position).Order("id DESC").Find(&ads)
	return ads
}
