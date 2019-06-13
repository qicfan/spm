package models

import (
	"github.com/jinzhu/gorm"
)

type UserCourseTest struct {
	gorm.Model
	UserID   uint   // 用户id
	CourseID uint   // 课程Id
	Duration int    // 持续时间，单位：秒
	Right    int    // 错误数量
	Wrong    int    // 正确数量
	Answers  string // 回答详情，json
}

type UserCourseAnswer struct {
	QuestionID uint   `json:"question_id"`
	Prefix     string `json:"prefix"`
}

func (UserCourseTest) TableName() string {
	return "user_course_test"
}
