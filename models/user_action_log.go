package models

import (
	"github.com/jinzhu/gorm"
	"supermentor/db"
	"supermentor/helpers"
	"time"
)

type UserActionType int

const (
	_                    UserActionType = iota
	UserOpenAppAction                   // 打开app 1
	UserLoginAction                     // 登录操作 2
	UserStudyActin                      // 学习课程操作 3
	UserInputDataAction                 // 记录数据操作 4
	UserUpdateInfoAction                // 修改用户信息 5
)

type UserActionLog struct {
	gorm.Model
	UserID     uint
	ActionType UserActionType
	ActionDate int64
	Remark     string
}

type ActionCount struct {
	ActionType  UserActionType
	ActionCount int
}

func (UserActionLog) TableName() string {
	return "user_action_log"
}

// 记录用户动作
func RecordUserAction(userID uint, actionType UserActionType, remark string) {
	actionLog := UserActionLog{
		UserID:     userID,
		ActionType: actionType,
		ActionDate: time.Now().Unix(),
		Remark:     remark,
	}
	if err := db.Db.Create(&actionLog).Error; err != nil {
		// 记录日志
		helpers.Log.Error("记录用户【%d】行为失败: %v", userID, err)
	}
}

// 查询用户一段时间内的动作次数汇总
func GetUserActionCount(userId uint, startTime int64, endTime int64) []ActionCount {
	actions := make([]ActionCount, 0)
	sql := "SELECT action_type, COUNT(id) as action_count FROM user_action_log WHERE user_id=? AND action_date>=? AND action_date<=? GROUP BY action_type"
	db.Db.Raw(sql, userId, startTime, endTime).Scan(&actions)
	return actions
}
