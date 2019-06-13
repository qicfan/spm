package models

import "github.com/jinzhu/gorm"

type TeamUser struct {
	gorm.Model
	TeamID uint // 团队id
	UserID uint // 成员Id
	RUser  User `gorm:"ForeignKey:UserID"` // 关联的用户
}

func (TeamUser) TableName() string {
	return "team_user"
}
