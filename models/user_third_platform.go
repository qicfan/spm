package models

import (
	"github.com/jinzhu/gorm"
)

type UserThirdPlatform struct {
	gorm.Model
	UserID         uint   // 用户 id
	Platform       string `gorm:"varchar(100);not null"` // 第三方平台类型，如微信
	PlatformUserId string `gorm:"varchar(200);not null"` // 第三方平台用户 id，如果 微信 openid
	Uniid          string // 微信用户统一ID
	Avatar         string `gorm:"varchar(500)"`
	City           string `gorm:"varchar(200)"`
	Province       string `gorm:"varchar(200)"`
	Country        string `gorm:"varchar(200)"`
	Nickname       string `gorm:"varchar(200)"` // 第三方平台的用户昵称
}

// 表名
func (UserThirdPlatform) TableName() string {
	return "user_third_platform"
}
