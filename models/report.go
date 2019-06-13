package models

import (
	"fmt"
	"github.com/go-redis/redis"
	"github.com/jinzhu/gorm"
	"sort"
	"supermentor/db"
	"supermentor/helpers"
	"time"
)

type ReportType int

const (
	_               ReportType = iota
	DayReportType              // 日报 1
	WeekReportType             // 周报 2
	MonthReportType            // 月报 3
	YearReportType             // 年报 4
)

type ReportStatus int
type ReportDatasSlice []ReportData

const (
	_                 ReportStatus = iota
	ReportStatusOpen               // 进行中 1
	ReportStatusClose              // 已计算 2
)

// 报告
type Report struct {
	gorm.Model
	UserID         uint         // 用户id
	ReportType     ReportType   // 报告类型
	Key            string       // 标识
	ReportDate     int64        // 报告时间戳，定位到当天零点零分零秒
	LastUpdateTime int64        // 最后更新时间
	DiligentNum    int          // 勤奋值
	Status         ReportStatus // 状态
	MaxUser        int          // 记录参与勤奋值排名的最大人数
	Rank           int          // 勤奋值排名

	ReportDatas []*ReportData
}

// 周报
type WeekReport struct {
	Week               *Report   // 周报
	Days               []*Report // 每日报告
	RecommendCourseIds []uint    // 推荐的课程
}

func (Report) TableName() string {
	return "report"
}

func (a ReportDatasSlice) Len() int { // 重写 Len() 方法
	return len(a)
}
func (a ReportDatasSlice) Swap(i, j int) { // 重写 Swap() 方法
	a[i], a[j] = a[j], a[i]
}
func (a ReportDatasSlice) Less(i, j int) bool { // 重写 Less() 方法， 从大到小排序
	return a[j].AbilityNum > a[i].AbilityNum
}

func (report *Report) GetVmParams(idKeys map[uint]DataColumn, idValues map[uint]float64) map[string]float64 {
	vmParams := make(map[string]float64)
	for id, dataColumn := range idKeys {
		// 存入虚拟机，准备开始计算
		vmParams[dataColumn.Key] = idValues[id]
		vmParams[dataColumn.Key+"Weighing"] = dataColumn.Weighing
		vmParams[dataColumn.Key+"DiligentWeighing"] = dataColumn.DiligentWeighing
		vmParams[dataColumn.Key+"AbilityWeighing"] = dataColumn.AbilityWeighing
		if report.ReportType == DayReportType {
			// 如果是日报，gama值除以7
			v := float64(dataColumn.GamaValue) / 7.00
			vmParams[dataColumn.Key+"Gama"] = v
		} else {
			vmParams[dataColumn.Key+"Gama"] = float64(dataColumn.GamaValue)
		}
	}
	return vmParams
}

// 查询或者生成一个用户的报告
func GetOrCreateReport(userId uint, reportType ReportType, key string, date time.Time) (*Report, error) {
	report := Report{
		UserID:     userId,
		ReportType: reportType,
		Key:        key,
	}
	if e := db.Db.Where(&report).Preload("ReportDatas").First(&report).RecordNotFound(); e {
		// 如果不存在，则创建一个
		report.UserID = userId
		report.ReportType = reportType
		report.Key = key
		report.DiligentNum = 0
		report.LastUpdateTime = time.Now().Unix()
		report.Status = ReportStatusOpen
		report.ReportDate = date.Unix()
		if err := db.Db.Create(&report).Error; err != nil {
			// 创建失败，返回错误
			helpers.Log.Error("%v\n\n", err)
			return &report, err
		}
	}
	return &report, nil
}

// 给报告添加或者更新某个数据项
func (report *Report) UpdateOrAddData(dataColumnId uint, dataValue float64) bool {
	var reportDataItem = ReportData{
		UserID:       report.UserID,
		ReportID:     report.ID,
		DataColumnID: dataColumnId,
	}
	dataIndex := -1
	for i := 0; i < len(report.ReportDatas); i++ {
		if report.ReportDatas[i].DataColumnID == dataColumnId && report.ReportDatas[i].Value == dataValue {
			// 没有变化，不需要更新
			helpers.Log.Info("%d的值没有变化，不需要更新", dataColumnId)
			return true
		}
		if report.ReportDatas[i].DataColumnID == dataColumnId {
			// 有变化，并且存在，记录索引
			dataIndex = i
			helpers.Log.Info("%d的值有变化，记录索引: %d", dataColumnId, dataIndex)
		}
	}
	tx := db.Db.Begin()
	if err := tx.Table("report_data").Where("report_id=? AND data_column_id=?", report.ID, dataColumnId).Update("value", dataValue).RowsAffected; err == 0 {
		// 更新失败
		helpers.Log.Error("更新%d的值失败：%v", dataColumnId, err)
		if tx.Error != nil {
			// 如果是数据库错误，则报错
			tx.Rollback()
			return false
		}
		reportDataItem.Value = dataValue
		if err := tx.Create(&reportDataItem).Error; err != nil {
			// 添加失败
			helpers.Log.Error("创建%d: %v的记录失败：%v", dataColumnId, dataValue, err)
			tx.Rollback()
			return false
		}
	}
	helpers.Log.Info("更新%d的值成功：%v", dataColumnId, tx.Error)
	tx.Commit()
	// 覆盖report.ReportDatas
	if dataIndex == -1 {
		// 不存在，则加入
		report.ReportDatas = append(report.ReportDatas, &reportDataItem)
	} else {
		report.ReportDatas[dataIndex].Value = dataValue
	}
	return true
}

func (report *Report) GetReportDataGroupByDataId() map[uint]float64 {
	// 将报告中的数据组成和dataId => value的字段
	idValues := make(map[uint]float64)
	for i := 0; i < len(report.ReportDatas); i++ {
		d := report.ReportDatas[i]
		idValues[d.DataColumnID] = d.Value
	}
	return idValues
}

// 计算报告的勤奋值，在更新数据后触发
func (report *Report) MakeReportDiligent() bool {
	// 引入公式，然后计算
	// 新增准客户数/22*0.15
	// +预约数/25*0.05
	// +确定预约数/8*0.05
	// +初次拜访数/8+后续拜访数/7）*0.2
	// +完成需求面谈数/4*0.1
	// +递送建议书数/5*0.05
	// +成交面谈数/5*0.05
	// +转介绍客户数/11*0.15
	// +(预约数/(初次拜访数+后续拜访数))/(5/3)*0.05
	// +((初次拜访数+后续拜访数)/完成需求面谈数)/(15/4)*0.05
	vmParams := report.GetVmParams(GetAllDataColumn(), report.GetReportDataGroupByDataId())
	formula := `(xkh / xkhGama * xkhWeighing) + (yy / yyGama * yyWeighing)+ (qdyy / qdyyGama * qdyyWeighing) + (ccbf/ccbfGama + hxbf/hxbfGama) * (ccbfWeighing + hxbfWeighing)+ (wzxqmt/wzxqmtGama * wzxqmtWeighing)+ (dsjys/dsjysGama * dsjysWeighing)+ (cjmt/cjmtGama * cjmtWeighing)+ (zjskh/zjskhGama * zjskhWeighing)`
	fmt.Printf("勤奋值公式：%v\n\n", formula)
	d := helpers.ExecJsFormula(formula, vmParams)
	if d == 0 {
		return false
	}
	// 更新勤奋值
	diligentNum := int(d * 100)
	fmt.Printf("diligentNum = : %v\n\n", diligentNum)
	report.DiligentNum = diligentNum
	if dbErr := db.Db.Table("report").Where("id = ?", report.ID).Update("diligent_num", diligentNum).Error; dbErr != nil {
		// 更新失败
		helpers.Log.Error("更新勤奋值失败：%v\n\n", dbErr)
		return false
	}
	return true
}

// 更新拜访量排名
// 两个拜访量相加，写入redis
func (report *Report) UpdateCallRank() bool {
	// 超过9点则不处理
	t := time.Now()
	finishTime := time.Date(t.Year(), t.Month(), t.Day(), 21, 0, 0, 0, t.Location()).Unix()
	if finishTime < t.Unix() {
		// 超过9点
		helpers.Log.Info("更新用户【%d】的拜访量【%s】失败：超过结算时间，不予更新", report.UserID, report.Key)
		return false
	}
	// 取两个拜访的数据相加
	dataColumns := GetAllDataColumn()
	idValues := report.GetReportDataGroupByDataId()
	bfl := 0.00
	// 记录两个拜访的数据
	for id, dataColumn := range dataColumns {
		if dataColumn.Key == "ccbf" || dataColumn.Key == "hxbf" {
			// 取数据相加
			if value, exist := idValues[id]; exist {
				// 存在
				bfl += value
			}
		}
	}
	if bfl == 0 {
		// 如果是0，则不处理
		return false
	}
	todayKey := helpers.GetDayReportKey(time.Now())
	rankKey := "CallRank-" + todayKey
	if err := db.Redis.ZAdd(rankKey, redis.Z{
		Score:  bfl,
		Member: report.ID,
	}).Err(); err != nil {
		return false
	}
	helpers.Log.Info("更新用户的拜访量排名【%s】：【%d】", report.Key, bfl)
	return true
}

// 查询某一周的报告
func GetWeekReport(uid uint, date time.Time) WeekReport {
	// 查询一周的key值
	dayReports := make([]*Report, 0)
	days := helpers.GetPerDayTimeInWeek(date)
	for i := 0; i < len(days); i++ {
		day := days[i]
		key := helpers.GetDayReportKey(day)
		report, _ := GetOrCreateReport(uid, DayReportType, key, day)
		dayReports = append(dayReports, report)
	}
	// 查询周报
	weekKey := helpers.GetDayReportKey(days[6])
	weekReport, _ := GetOrCreateReport(uid, WeekReportType, weekKey, days[6])
	// 查询推荐的课程
	courseIds := GetCourseIdsByReportId(weekReport.ID)
	week := WeekReport{
		Week:               weekReport,
		Days:               dayReports,
		RecommendCourseIds: courseIds,
	}
	return week
}

// 更新周报的数据，将每日单项数据相加然后更新；计算一周勤奋值
func (weekReport *WeekReport) UpdateWeekReportData() bool {
	datas := make(map[uint]float64)
	for i := 0; i < len(weekReport.Days); i++ {
		day := weekReport.Days[i]
		for x := 0; x < len(day.ReportDatas); x++ {
			d := day.ReportDatas[x]
			dv, e := datas[d.DataColumnID]
			if !e {
				// 不存在，则赋值
				datas[d.DataColumnID] = d.Value
			} else {
				datas[d.DataColumnID] = dv + d.Value
			}
		}
	}
	fmt.Printf("加和后的周报数据：%+v\n\n", datas)
	// 更新数据后入库
	for k, v := range datas {
		fmt.Printf("更新 %d 的值为: %v\n", k, v)
		weekReport.Week.UpdateOrAddData(k, v)
	}
	// 计算勤奋值
	weekReport.Week.MakeReportDiligent()
	return true
}

// 更新周报的单项能力得分，用来推荐课程和生成能力雷达图，每周执行一次
func (weekReport *WeekReport) UpdateWeekReportRecommendNum() bool {
	idColumns := GetAllDataColumn()
	vmParams := weekReport.Week.GetVmParams(idColumns, weekReport.Week.GetReportDataGroupByDataId())
	for i := 0; i < len(weekReport.Week.ReportDatas); i++ {
		d := weekReport.Week.ReportDatas[i]
		// 取得公式
		formula := idColumns[d.DataColumnID].RecommendFormula
		if formula == "" {
			continue
		}
		result := helpers.ExecJsFormula(formula, vmParams)
		if result == 0 {
			continue
		}
		d.AbilityNum = result
		// 更新到数据库
		db.Db.Table("report_data").Where("id=?", d.ID).Update("ability_num", result)
		if db.Db.Error != nil {
			// 没有更新，记录日志
			helpers.Log.Warning("更新周报【%s】的数据【%s】的能力值失败", weekReport.Week.Key, idColumns[d.DataColumnID].Name)
		}
	}
	return true
}

// 更新周报的推荐课程，每周执行一次
// 根据能力值最低的两个能立项推荐对应的课程，已经推荐过的不推
func (weekReport *WeekReport) UpdateWeekReportRecommendCourse() bool {
	l := len(weekReport.Week.ReportDatas)
	if l == 0 {
		return false
	}
	dataColumns := GetAllDataColumn()
	// 排除掉0值的数据，然后进行排序
	excludeDataId := make(map[uint]uint)
	for k, v := range dataColumns {
		if v.AbilityID == 0 {
			excludeDataId[k] = 1
		}
	}
	// 进行本周能力项排序，进行一次
	rd := make([]ReportData, 0)
	for i := 0; i < l; i++ {
		if _, exist := excludeDataId[weekReport.Week.ReportDatas[i].DataColumnID]; exist {
			// 如果这个数据项没有对应能力就跳过
			continue
		}
		rd = append(rd, *weekReport.Week.ReportDatas[i])
	}
	sort.Sort(ReportDatasSlice(rd))
	// 取dataColumn
	n := 0
	for i := 0; i < len(rd); i++ {
		helpers.Log.Info("第%d次处理课程", n)
		if n >= 2 {
			break
		}
		aId := dataColumns[rd[i].DataColumnID].AbilityID
		// 取能力对应的课程，选一个
		// 取用户该能力已推荐的课程，做排除
		existCourseIds := GetUserCourseIDsByAbilityID(aId, weekReport.Week.UserID)
		// 取能力对应的课程ids
		recommendCourses := GetCoursesByAbilityID(aId, existCourseIds)
		if len(recommendCourses) == 0 {
			helpers.Log.Info("能立项【%d】没有对应的课程", aId)
			continue
		}
		// 写入推荐课程
		for x := 0; x < len(recommendCourses); x++ {
			if x >= 2 {
				break
			}
			weekReport.AddReportCourse(recommendCourses[x], aId)
		}
		n++
	}
	helpers.Log.Info("周报推荐课程完毕:%d", n)
	return true
}

// 结算周勤奋值排名，从redis写入数据库
func CloseWeekReportDiligentRank(key string) {
	helpers.Log.Info("开始处理周报的勤奋值排名")
	// 取到data_column_id
	rankKey := "DiligentRank-" + key
	// 从redis中取有序集合，每次取10个
	b := 0
	c := 0
	zSetCount, _ := db.Redis.ZCard(rankKey).Result()
	count := int(zSetCount)
	helpers.Log.Info("查询参与周报排名的总人数:%d", count)
	for {
		stop := false
		start := (b - 1) * 100
		end := start + 99
		result, _ := db.Redis.ZRevRange(rankKey, int64(start), int64(end)).Result()
		helpers.Log.Info("第%d页有%d个结果", b, len(result))
		if len(result) == 0 {
			break
		}
		for i := 0; i < len(result); i++ {
			reportId := result[i]
			// 直接更新
			db.Db.Table("report").Where("id=?", reportId).Update(map[string]interface{}{
				"max_user": count,
				"rank":     c,
			})
			c++
			if c > count {
				stop = true
				break
			}
		}
		if stop {
			break
		}
		b++
	}
	helpers.Log.Info("周报的勤奋值排名处理完毕")
}

// 结算周报，计算能力度分数、推荐课程、修改状态为Close
func (weekReport *WeekReport) Close() {
	if weekReport.Week.Status == ReportStatusClose {
		return
	}
	weekReport.UpdateWeekReportRecommendNum()
	weekReport.UpdateWeekReportRecommendCourse()
	weekReport.Week.Status = ReportStatusClose
	db.Db.Model(weekReport.Week).Update("status", ReportStatusClose)
}

func (weekReport *WeekReport) TodayReport(t time.Time) *Report {
	todayKey := helpers.GetDayReportKey(t)
	for i := 0; i < len(weekReport.Days); i++ {
		if weekReport.Days[i].Key == todayKey {
			return weekReport.Days[i]
		}
	}
	helpers.Log.Error("没有找到今日报告: %s, %+v", todayKey, weekReport.Days)
	return nil
}

// 给报告添加课程
func (weekReport *WeekReport) AddReportCourse(course Course, abilityId uint) {
	reportCourse := ReportCourse{
		CourseID:  course.ID,
		AbilityID: abilityId,
		UserID:    weekReport.Week.UserID,
		ReportID:  weekReport.Week.ID,
	}
	userCourse := UserCourse{
		CourseID:  course.ID,
		AbilityID: abilityId,
		UserID:    weekReport.Week.UserID,
	}
	if err := db.Db.Create(&reportCourse).Error; err != nil {
		// 添加失败,记录日志
		helpers.Log.Error("报告【%s】添加课程【%d】失败：%v", weekReport.Week.Key, course.ID, err)
	}
	if err := db.Db.Create(&userCourse).Error; err != nil {
		// 添加失败,记录日志
		helpers.Log.Error("用户【%d】添加课程【%d】失败：%v", weekReport.Week.UserID, course.ID, err)
	}
	// 增加学习中的人数
	course.IncreaseStudyUser()
	l := len(weekReport.RecommendCourseIds)
	exist := false
	// 查询是否已经存在
	for i := 0; i < l; i++ {
		if weekReport.RecommendCourseIds[i] == course.ID {
			exist = true
			break
		}
	}
	if !exist {
		// 不存在，则添加
		weekReport.RecommendCourseIds = append(weekReport.RecommendCourseIds, course.ID)
	}
}

// 更新周拜访量排名
func (weeReport *WeekReport) UpdateDiligentRank() bool {
	rankKey := "DiligentRank-" + weeReport.Week.Key
	if err := db.Redis.ZAdd(rankKey, redis.Z{
		Score:  float64(weeReport.Week.DiligentNum),
		Member: weeReport.Week.ID,
	}).Err(); err != nil {
		return false
	}
	helpers.Log.Info("更新周报的拜访量排名【%s】：【%d】", weeReport.Week.Key, weeReport.Week.DiligentNum)
	return true
}

// 结算本周的周报
func CloseWeekReport() {
	// 取本周未结算的周报
	dayKeys := helpers.GetPerDayTimeInWeek(time.Now())
	weekTime := dayKeys[6]
	//weekTime := time.Date(2018, 4, 15, 0, 0, 0, 0, time.Now().Location())
	key := helpers.GetDayReportKey(weekTime)
	reports := make([]Report, 0)
	// 查询有数据的用户
	sql := "SELECT user_id FROM `report` WHERE `key`=? AND report_type=? AND status=?"
	db.Db.Raw(sql, key, WeekReportType, ReportStatusOpen).Scan(&reports)
	for _, report := range reports {
		// 查询周报
		weekReport := GetWeekReport(report.UserID, weekTime)
		helpers.Log.Info("用户【%d】的周报【%s】开始结算", report.UserID, key)
		weekReport.Close()
		helpers.Log.Info("用户【%d】的周报【%s】已结算", report.UserID, key)
	}
	CloseWeekReportDiligentRank(key)
}

// 结算每日的拜访量排行榜
// 将数据从redis取出，更新到数据库
func CloseDayReportCallRank() {
	// 取到data_column_id
	var dataColumnId uint
	dataColumns := GetAllDataColumn()
	for id, dataColumn := range dataColumns {
		if dataColumn.Key == "hxbf" {
			dataColumnId = id
			break
		}
	}
	todayKey := helpers.GetDayReportKey(time.Now())
	rankKey := "CallRank-" + todayKey
	// 从redis中取有序集合，每次取10个
	b := 0
	c := 0
	zSetCount, _ := db.Redis.ZCard(rankKey).Result()
	count := int(zSetCount)
	for {
		stop := false
		start := (b - 1) * 100
		end := start + 99
		result, _ := db.Redis.ZRevRange(rankKey, int64(start), int64(end)).Result()
		for i := 0; i < len(result); i++ {
			reportId := result[i]
			// 直接更新
			db.Db.Table("report_data").Where("report_id=? AND data_column_id=?", reportId, dataColumnId).Update(map[string]interface{}{
				"max_user": count,
				"rank":     c,
			})
			c++
			if c > count {
				stop = true
				break
			}
		}
		if stop {
			break
		}
		b++
	}
}

// 查询用户的报告，每页30条，倒序排列
func GetUserReport(userId uint, page int) []Report {
	reports := make([]Report, 0)
	pageSize := 30
	now := time.Now().Unix()
	db.Db.Where("user_id=? AND report_type=? AND report_date<?", userId, DayReportType, now).Preload("ReportDatas").Order("report_date DESC, id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&reports)
	return reports
}
