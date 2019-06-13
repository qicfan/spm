package models

import (
	"github.com/jinzhu/gorm"
	"supermentor/db"
	"supermentor/helpers"
)

type Course struct {
	gorm.Model
	Title              string  `gorm:"type:varchar(200)"` // 课程名称
	Desc               string  `gorm:"type:varchar(500)"` // 描述
	Content            string  `gorm:"type:mediumtext"`   // 内容
	Cover              string  `gorm:"type:varchar(500)"` // 封面图片
	Audio              string  `gorm:"type:varchar(200)"` // 音频链接
	AudioDuration      float64 // 音频时长，秒
	TotalUser          int     // 学完的人
	StudyUser          int     // 正在学的人
	StandardTrainAudio string  // 训练标准音频
	AbilityID          uint    `gorm:"type:int"` // 课程对应的能立项
	Tests              []*CourseTest
}

func (Course) TableName() string {
	return "course"
}

// 根据id查询课程详情
func GetCourseById(courseId uint) *Course {
	course := &Course{}
	db.Db.Preload("Tests").Find(course, courseId)
	return course
}

// 取能力项对应的课程ids，并且排除掉指定的课程
func GetCoursesByAbilityID(abilityID uint, excludeIds []uint) []Course {
	courses := make([]Course, 0)
	if len(excludeIds) == 0 {
		db.Db.Where("ability_id=?", abilityID).Find(&courses)
	} else {
		db.Db.Where("id NOT IN (?) AND ability_id=?", excludeIds, abilityID).Find(&courses)
	}
	return courses
}

// 增加完成学习的人数
func (course *Course) IncreaseTotalUser() {
	if err := db.Db.Model(course).UpdateColumn("total_user", gorm.Expr("total_user + ?", 1)); err != nil {
		helpers.Log.Error("课程【%d】增加完成学习的人数失败:%v", course.ID, err)
	}
}

// 减少学习中的人数
func (course *Course) DecreaseStudyUser() {
	if err := db.Db.Model(course).UpdateColumn("study_user", gorm.Expr("study_user - ?", 1)); err != nil {
		helpers.Log.Error("课程【%d】减少学习中的人数失败:%v", course.ID, err)
	}
}

// 增加学习中的人数
func (course *Course) IncreaseStudyUser() {
	if err := db.Db.Model(course).UpdateColumn("study_user", gorm.Expr("study_user + ?", 1)); err != nil {
		helpers.Log.Error("课程【%d】增加学习中的人数失败:%v", course.ID, err)
	}
}

// 添加或更新课程
func (course *Course) Save() error {
	if db.Db.NewRecord(course) {
		// 新增课程
		err := db.Db.Create(course).Error
		return err
	}
	return db.Db.Model(course).Update(course).Error
}

// 查询课程列表
func GetCourseList(page int) ([]Course, int) {
	courses := make([]Course, 0)
	total := 0
	pageSize := 100
	db.Db.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&courses)
	db.Db.Model(&Course{}).Count(&total)
	return courses, total
}

// 查询学习该课程的用户列表
func (course *Course) GetCourseStudyUsers() []*UserCourse {
	userCourses := make([]*UserCourse, 0)
	db.Db.Where("course_id=? AND finish_time=0", course.ID).Preload("RUser").Find(&userCourses)
	return userCourses
}
