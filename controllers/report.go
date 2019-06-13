package controllers

import (
	"encoding/json"
	"strconv"
	"supermentor/models"
	"time"
)

// 查询周报，包含每天
func GetWeekReport(route *SmRoute) {
	var day time.Time
	var userId uint
	r := route.Request
	dateString := r.Form.Get("date")
	if dateString == "" {
		// 如果没有参数，则默认取本周
		day = time.Now()
	} else {
		date, _ := strconv.ParseInt(dateString, 10, 64)
		day = time.Unix(date, 0)
	}
	userIdString := r.Form.Get("user_id")
	userIdInt, err := strconv.Atoi(userIdString)
	if err != nil {
		userId = uint(userIdInt)
	} else {
		userId = route.UserId
	}
	report := models.GetWeekReport(userId, day)
	route.ReturnJson(CODE_OK, "", report)
	return
}

// 查询所有数据字段
func GetReportDataColumn(route *SmRoute) {
	columns := models.GetAllDataColumnToArray()
	// 只有第一次访问app时会请求数据项
	go models.RecordUserAction(route.UserId, models.UserOpenAppAction, "打开APP")
	route.ReturnJson(CODE_OK, "", columns)
	return
}

// 更新日报数据
func UpdateData(route *SmRoute) {
	r := route.Request
	t := time.Now()
	tString := r.Form.Get("time")
	if tString != "" {
		tInt, tErr := strconv.ParseInt(tString, 10, 64)
		if tErr == nil {
			t = time.Unix(tInt, 0)
		}
	}
	closeWeekString := r.Form.Get("close_week")
	closeWeek := 0 // 是否关闭周报
	if closeWeekString == "1" {
		closeWeek = 1
	}
	// 查询出周报
	weekReport := models.GetWeekReport(route.UserId, t)
	todayReport := weekReport.TodayReport(t)
	if todayReport == nil {
		// 有错误
		route.ReturnJson(CODE_ERROR, "今日报告不存在", "")
		return
	}
	// 得到客户端上传的数据{id: value}
	dataJson := r.Form.Get("data")
	data := make(map[int]float64)
	// 数据数据
	json.Unmarshal([]byte(dataJson), &data)
	for dataId, v := range data {
		todayReport.UpdateOrAddData(uint(dataId), v)
	}
	// 更新日报勤奋值
	todayReport.MakeReportDiligent()
	// 更新周报数据，然后更新周勤奋值
	weekReport.UpdateWeekReportData()
	// 更新今日拜访量排名
	go todayReport.UpdateCallRank()
	// 更新周勤奋值排名
	go weekReport.UpdateDiligentRank()
	go models.RecordUserAction(route.UserId, models.UserInputDataAction, "录入数据")
	if closeWeek == 1 {
		// 关闭周报
		weekReport.Close()
	}
	route.ReturnJson(CODE_OK, "", weekReport)
	return
}

// 查询用户周报推荐的课程
func GetReportUserCourse(route *SmRoute) {
	reportIdString := route.Request.Form.Get("report_id")
	if reportIdString == "" {
		route.ReturnJson(CODE_ERROR, "参数错误", "")
	}
	reportId, _ := strconv.ParseInt(reportIdString, 10, 64)
	// 查询报告对应的用户课程
	userCourses := models.GetUserCourseByReportId(uint(reportId))
	route.ReturnJson(CODE_OK, "", userCourses)
}
