package models

import (
	"github.com/jinzhu/gorm"
	"supermentor/db"
)

type ReportCourse struct {
	gorm.Model
	UserID    uint
	ReportID  uint
	CourseID  uint
	AbilityID uint
}

func (ReportCourse) TableName() string {
	return "report_course"
}

// 查询报告推荐的课程
func GetCourseIdsByReportId(reportId uint) []uint {
	var reportCourses []ReportCourse
	courseIds := make([]uint, 1)
	if db.Db.Where("report_id = ?", reportId).Find(&reportCourses).RecordNotFound() {
		// 找不到
		return courseIds
	}
	for i := 0; i < len(reportCourses); i++ {
		courseIds = append(courseIds, reportCourses[i].CourseID)
	}
	return courseIds
}

// 查询报告对应的用户课程
func GetUserCourseByReportId(reportId uint) []*UserCourse {
	reportCourses := make([]ReportCourse, 0)
	userCourses := make([]*UserCourse, 0)

	if db.Db.Where("report_id = ?", reportId).Find(&reportCourses).RecordNotFound() {
		return userCourses
	}

	for i := 0; i < len(reportCourses); i++ {
		userCourse := GetUserCourseByID(reportCourses[i].CourseID, reportCourses[i].UserID)
		userCourses = append(userCourses, userCourse)
	}
	return userCourses
}
