package models

import (
	"encoding/json"
	"github.com/jinzhu/gorm"
	"supermentor/db"
	"supermentor/helpers"
	"time"
)

type UserCourse struct {
	gorm.Model
	UserID     uint  // 用户id
	AbilityID  uint  // 能力值id
	CourseID   uint  // 课程id
	Progress   int   // 学习进度，完成为100
	StartTime  int64 // 开始学习时间, 时间戳
	FinishTime int64 // 完成学习时间, 时间戳
	TrainCount int   // 训练次数
	TestResult int   // 自测结果

	RCourse Course `gorm:"ForeignKey:CourseID"` // 关联的课程
	RUser   User   `gorm:"ForeignKey:UserID"`   // 关联的用户
}

func (UserCourse) TableName() string {
	return "user_course"
}

func GetUserCourseByID(courseId uint, userId uint) *UserCourse {
	userCourse := &UserCourse{}
	db.Db.Preload("RCourse").Where("course_id=? AND user_id=?", courseId, userId).Find(userCourse)
	return userCourse
}

// 查询用户某个能立项推荐的课程
func GetUserCourseIDsByAbilityID(abilityID uint, userID uint) []uint {
	courseIds := make([]uint, 0)
	courses := make([]UserCourse, 0)
	db.Db.Where("ability_id=? AND user_id=?", abilityID, userID).Find(&courses)
	if db.Db.Error != nil {
		// 有错误，返回空
		return courseIds
	}
	for i := 0; i < len(courses); i++ {
		courseIds = append(courseIds, courses[i].CourseID)
	}
	return courseIds
}

// 查询推荐给用户的课程，返回课程列表
func GetUserRecommendCourse(uid uint, isFinish int) []UserCourse {
	courses := make([]UserCourse, 0)
	// 查询最近推荐的两个课程
	if isFinish == 1 {
		db.Db.Where("user_id=? AND finish_time>1", uid).Order("id DESC").Preload("RCourse").Find(&courses)
	} else {
		db.Db.Where("user_id=? AND finish_time=0", uid).Order("id DESC").Preload("RCourse").Find(&courses)
	}
	return courses
}

// 记录开始学习课程的时间
func (userCourse *UserCourse) Start(time int64) bool {
	if userCourse.StartTime > 0 {
		helpers.Log.Warning("更新课程【%d】的开始时间【%d】失败：课程已经开始学习", userCourse.ID, time)
		return false
	}
	if row := db.Db.Model(userCourse).Update("start_time", time).RowsAffected; row == 0 {
		// 如果没更新，则记录日志
		helpers.Log.Warning("更新课程【%d】的开始时间【%d】失败", userCourse.ID, time)
	}
	return true
}

// 更新学习进度
func (userCourse *UserCourse) UpdateProgress(progress int) bool {
	if userCourse.FinishTime > 0 {
		helpers.Log.Warning("课程已经学习完成，不会再更新", userCourse.ID, progress, userCourse.Progress)
		return false
	}
	if progress <= userCourse.Progress {
		helpers.Log.Warning("更新课程【%d】的进度【%d】失败：新的进度小于等于老就进度%d", userCourse.ID, progress, userCourse.Progress)
		return false
	}
	if row := db.Db.Model(userCourse).Update("progress", progress).RowsAffected; row == 0 {
		// 如果没更新，则记录日志
		helpers.Log.Warning("更新课程【%d】的进度【%d】失败", userCourse.ID, progress)
		return false
	}
	return true
}

// 完成学习，记录完成时间，改变状态
func (userCourse *UserCourse) Finish() bool {
	userCourse.FinishTime = time.Now().Unix()
	if err := db.Db.Model(&userCourse).Update("finish_time", userCourse.FinishTime); err != nil {
		helpers.Log.Warning("用户【%d】完成学习课程【%d】失败", userCourse.UserID, userCourse.ID, userCourse.FinishTime)
		return false
	}
	userCourse.RCourse.IncreaseTotalUser()
	userCourse.RCourse.DecreaseStudyUser()
	return true
}

// 查询课程的自测列表
func (userCourse *UserCourse) GetUserCourseTests() []UserCourseTest {
	list := make([]UserCourseTest, 0)
	db.Db.Order("id DESC").Where("user_id=? AND course_id=?", userCourse.UserID, userCourse.CourseID).Find(&list)
	return list
}

// 保存自测结果
// answers是问题回答详情，[{questionID: answerPrefix}]的结果
func (userCourse *UserCourse) SaveTest(duration int, answers string) *UserCourseTest {
	var right = 0
	var wrong = 0
	test := &UserCourseTest{
		UserID:   userCourse.UserID,
		CourseID: userCourse.CourseID,
		Duration: duration,
		Answers:  answers,
	}
	answersDecoded := make([]UserCourseAnswer, 0)
	// 解开answers，计算right & wrong
	err := json.Unmarshal([]byte(answers), &answersDecoded)
	if err != nil {
		// 有错误
		helpers.Log.Error("用户【%d】课程【%d】自测保存失败，解析问题答案出错：%v", userCourse.UserID, userCourse.CourseID, err)
		return test
	}
	// 取所有问题
	rightAnswers := make(map[uint]string, 0)
	questions := GetCourseTest(userCourse.CourseID)
	for i := 0; i < len(questions); i++ {
		rightAnswers[questions[i].ID] = questions[i].Answer
	}
	l := len(answersDecoded)
	for i := 0; i < l; i++ {
		if answersDecoded[i].Prefix == rightAnswers[answersDecoded[i].QuestionID] {
			// 正确
			right++
		} else {
			wrong++
		}
	}
	test.Right = right
	test.Wrong = wrong
	// 入库
	if err := db.Db.Create(test).Error; err != nil {
		// 出错
		helpers.Log.Error("用户【%d】课程【%d】自测保存失败，保存答案到数据库失败:%v", userCourse.UserID, userCourse.CourseID, err)
	}
	if test.Wrong == 0 {
		// 全对，则完成学习
		userCourse.Finish()
	}
	return test
}
