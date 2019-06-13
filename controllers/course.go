package controllers

import (
	"encoding/json"
	"strconv"
	"supermentor/db"
	"supermentor/models"
)

// 查询课程详情
func GetCourseDetail(route *SmRoute) {
	courseIdString := route.Request.Form.Get("id")
	if courseIdString == "" {
		route.ReturnJson(CODE_ERROR, "参数错误", "")
		return
	}
	courseId, _ := strconv.Atoi(courseIdString)
	// 查询课程
	course := models.GetCourseById(uint(courseId))
	route.ReturnJson(CODE_OK, "", course)
}

// 更新课程学习进度，满值100
func UpdateCourseProgress(route *SmRoute) {
	courseIdString := route.Request.Form.Get("courseId")
	progressString := route.Request.Form.Get("progress")
	if courseIdString == "" {
		route.ReturnJson(CODE_ERROR, "参数错误", "")
		return
	}
	courseId, _ := strconv.Atoi(courseIdString)
	progress, _ := strconv.Atoi(progressString)
	// 查找course
	userCourse := models.GetUserCourseByID(uint(courseId), route.UserId)
	if userCourse.ID == 0 {
		route.ReturnJson(CODE_ERROR, "课程不存在", "")
		return
	}
	userCourse.UpdateProgress(progress)
	go models.RecordUserAction(route.UserId, models.UserStudyActin, "学习课程")
	route.ReturnJson(CODE_OK, "", "")
}

// 更新课程开始学习时间
func StartCourse(route *SmRoute) {
	courseIdString := route.Request.Form.Get("courseId")
	timeString := route.Request.Form.Get("time")
	if courseIdString == "" {
		route.ReturnJson(CODE_ERROR, "参数错误", "")
		return
	}
	courseId, _ := strconv.Atoi(courseIdString)
	time, _ := strconv.Atoi(timeString)
	// 查找course
	userCourse := models.GetUserCourseByID(uint(courseId), route.UserId)
	if userCourse.ID == 0 {
		route.ReturnJson(CODE_ERROR, "课程不存在", "")
		return
	}
	userCourse.Start(int64(time))
	route.ReturnJson(CODE_OK, "", "")
}

// 查询课程考核题目
func GetCourseQuestion(route *SmRoute) {
	courseIdString := route.Request.Form.Get("course_id")
	if courseIdString == "" {
		route.ReturnJson(CODE_ERROR, "参数错误", "")
		return
	}
	courseId, _ := strconv.Atoi(courseIdString)
	tests := models.GetCourseTest(uint(courseId))
	route.ReturnJson(CODE_OK, "", tests)
}

// 保存考核结果
func SaveTestResult(route *SmRoute) {
	courseIdString := route.Request.Form.Get("course_id")
	answers := route.Request.Form.Get("answers")
	durationString := route.Request.Form.Get("duration")
	if courseIdString == "" || answers == "" || durationString == "" {
		route.ReturnJson(CODE_ERROR, "参数错误", "")
		return
	}
	courseId, courseIdErr := strconv.Atoi(courseIdString)
	duration, durationErr := strconv.Atoi(durationString)
	if courseIdErr != nil || durationErr != nil {
		route.ReturnJson(CODE_ERROR, "参数非法", "")
		return
	}
	course := models.GetCourseById(uint(courseId))
	if course.ID == 0 {
		route.ReturnJson(CODE_ERROR, "课程不存在", "")
		return
	}
	userCourse := models.GetUserCourseByID(uint(courseId), route.UserId)
	if userCourse.ID == 0 {
		route.ReturnJson(CODE_ERROR, "您还没有学习该课程", "")
		return
	}
	test := userCourse.SaveTest(duration, answers)
	route.ReturnJson(CODE_OK, "", test)
	return
}

// 查询考核列表
func GetCourseTests(route *SmRoute) {
	courseIdString := route.Request.Form.Get("course_id")
	if courseIdString == "" {
		route.ReturnJson(CODE_ERROR, "参数错误", "")
		return
	}
	courseId, courseIdErr := strconv.Atoi(courseIdString)
	if courseIdErr != nil {
		route.ReturnJson(CODE_ERROR, "参数非法", "")
		return
	}
	userCourse := models.GetUserCourseByID(uint(courseId), route.UserId)
	if userCourse.ID == 0 {
		route.ReturnJson(CODE_ERROR, "您还没有学习该课程", "")
		return
	}
	tests := userCourse.GetUserCourseTests()
	route.ReturnJson(CODE_OK, "", tests)
}

// 查询正在学习该课程的前几个用户，返回id, name, avatar
func GetCourseStudyUsers(route *SmRoute) {
	courseIdString := route.Request.Form.Get("course_id")
	if courseIdString == "" {
		route.ReturnJson(CODE_ERROR, "参数错误", "")
		return
	}
	courseId, courseIdErr := strconv.Atoi(courseIdString)
	if courseIdErr != nil {
		route.ReturnJson(CODE_ERROR, "参数非法", "")
		return
	}
	course := models.GetCourseById(uint(courseId))
	if course.ID == 0 {
		route.ReturnJson(CODE_ERROR, "课程不存在", "")
		return
	}
	userCourses := course.GetCourseStudyUsers()
	l := len(userCourses)
	if l == 0 {
		route.ReturnJson(CODE_OK, "", userCourses)
		return
	}
	users := make([]models.User, 0)
	for i := 0; i < l; i++ {
		userCourses[i].RUser.Password = ""
		userCourses[i].RUser.Mobile = ""
		users = append(users, userCourses[i].RUser)
	}
	route.ReturnJson(CODE_OK, "", users)
	return
}

// 后台保存自测题目
func SaveCourseQuestion(route *SmRoute) {
	courseIdString := route.Request.Form.Get("course_id")
	if courseIdString == "" {
		route.ReturnJson(CODE_ERROR, "参数错误", "")
		return
	}
	courseId, courseIdErr := strconv.Atoi(courseIdString)
	if courseIdErr != nil {
		route.ReturnJson(CODE_ERROR, "参数非法", "")
		return
	}
	course := models.GetCourseById(uint(courseId))
	if course.ID == 0 {
		route.ReturnJson(CODE_ERROR, "课程不存在", "")
		return
	}
	data := route.Request.Form.Get("data")
	questions := make([]models.CourseTest, 0)
	_ = json.Unmarshal([]byte(data), &questions)
	ids := make([]interface{}, 0)
	for i := 0; i < len(questions); i++ {
		items := questions[i].Items
		answers, _ := json.Marshal(items)
		questions[i].Answers = string(answers[:])
		// 保存到数据库
		if questions[i].ID > 0 {
			// 更新
			ids = append(ids, questions[i].ID)
			db.Db.Model(&models.CourseTest{}).Update(questions[i])
		} else {
			// 新增
			db.Db.Create(&questions[i])
			if questions[i].ID > 0 {
				ids = append(ids, questions[i].ID)
			}
		}
	}
	// 删除不存在的
	db.Db.Where("id NOT IN(?) AND course_id=?", ids, courseId).Delete(&models.CourseTest{})
	route.ReturnJson(CODE_OK, "", questions)
	return
}
