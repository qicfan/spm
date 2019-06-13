package models

import (
	"github.com/jinzhu/gorm"
	"supermentor/db"
)

type CourseTest struct {
	gorm.Model
	CourseID uint
	Title    string
	pos      uint
	Answer   string
	Answers  string
	Items    []interface{} `gorm:"-"`
}

func (CourseTest) TableName() string {
	return "course_test"
}

// 查询课程自测的题目
func GetCourseTest(courseId uint) []CourseTest {
	courseTests := make([]CourseTest, 0)
	db.Db.Where("course_id=?", courseId).Order("pos ASC").Find(&courseTests)
	return courseTests
}
