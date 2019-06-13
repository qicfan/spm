package models

import (
	"github.com/jinzhu/gorm"
	"supermentor/db"
)

type TeamApplyStatus int

const (
	_                     TeamApplyStatus = iota
	TeamApplyStatusWait                   // 待审核 1
	TeamApplyStatusPass                   // 审核通过 2
	TeamApplyStatusReject                 // 审核驳回 3
)

type TeamApply struct {
	gorm.Model
	TeamID uint            // 团队id
	UserID uint            // 成员Id
	Status TeamApplyStatus // 状态

	ApplyUser *User `gorm:"ForeignKey:UserID"` // 申请人
	ApplyTeam *Team `gorm:"ForeignKey:TeamID"` // 申请加入的团队
}

func (TeamApply) TableName() string {
	return "team_apply"
}

// 查找团队加入申请
func GetTeamApply(user *User, team *Team) *TeamApply {
	ta := &TeamApply{
		TeamID: team.ID,
		UserID: user.ID,
	}
	db.Db.Where(ta).Find(ta)
	return ta
}

// 审核通过
func (ta *TeamApply) TeamApplyPass() {
	ta.Status = TeamApplyStatusPass
	db.Db.Model(&ta).Update("status", TeamApplyStatusPass)
}

// 审核驳回
func (ta *TeamApply) TeamApplyReject() {
	ta.Status = TeamApplyStatusReject
	db.Db.Model(&ta).Update("status", TeamApplyStatusReject)
}
