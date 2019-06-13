package models

import (
	"github.com/jinzhu/gorm"
	"supermentor/db"
	"supermentor/helpers"
	"time"
)

type Team struct {
	gorm.Model
	OwnerID   uint   // 团队长id
	Name      string // 团队名称，默认为团队长名字+的团队
	UserCount int    // 团队成员数量
	ParentID  uint   // 上级团队ID

	Owner     *User
	TeamUsers []*TeamUser // 团队所属成员
	Teams     []*Team     // 所辖团队
}

func (Team) TableName() string {
	return "team"
}

// 创建团队
func CreateTeam(user *User) *Team {
	team := &Team{
		OwnerID:   user.ID,
		Name:      user.Nickname + "的团队",
		UserCount: 1,
		ParentID:  user.BelongTeam.ID, // 父级Id等于用户所属团队的ID
	}
	db.Db.Create(team)
	user.MyTeam = team
	return team
}

// 通过团队id查询团队
func GetTeamById(id uint) *Team {
	team := &Team{}
	db.Db.Find(team, id)
	return team
}

// 通过团队长id查询团队
func GetOwnerTeam(userID uint) *Team {
	team := &Team{}
	db.Db.Where("owner_id=?", userID).Find(team)
	return team
}

// 查询用户所属团队
func GetUserTeamByUserID(userID uint) *Team {
	tu := &TeamUser{}
	team := &Team{}
	if db.Db.Where("user_id=?", userID).Find(tu).RecordNotFound() {
		return team
	}
	// 查询team
	db.Db.Find(team, tu.TeamID)
	return team
}

// 使用手机号查找团队
func SearchTeamByUserMobile(mobile string) *Team {
	// 通过手机号查询用户，通过用户查询团队
	user := GetUserByMobile(mobile, false)
	if user.ID == 0 {
		return &Team{}
	}
	team := GetOwnerTeam(user.ID)
	if team.ID > 0 {
		team.Owner = user
	}
	return team
}

// 减少成员数量
func (team *Team) DecreaseMemberCount(c int) {
	if err := db.Db.Model(team).UpdateColumn("user_count", gorm.Expr("user_count - ?", c)); err != nil {
		helpers.Log.Error("团队【%d】减少成员数量失败:%v", team.ID, err)
	}
	team.UserCount -= c
}

// 增加成员数量
func (team *Team) IncreaseMemberCount(c int) {
	if err := db.Db.Model(team).UpdateColumn("user_count", gorm.Expr("user_count + ?", c)); err != nil {
		helpers.Log.Error("团队【%d】增加成员数量失败:%v", team.ID, err)
	}
	team.UserCount += c
}

// 和上级团队脱离解除关系
func (team *Team) LeaveParent() {
	team.ParentID = 0
	db.Db.Model(team).Update("parent_id", 0)
}

// 添加一个团队加入申请
func (team *Team) TeamAddApply(user *User) bool {
	// 建立关系
	tu := &TeamApply{
		TeamID: team.ID,
		UserID: user.ID,
		Status: TeamApplyStatusWait,
	}
	if db.Db.Where(tu).Find(tu).RecordNotFound() {
		if err := db.Db.Create(tu).Error; err != nil {
			helpers.Log.Error("团队【%d】增加成员【%d】失败:%v", team.ID, user.ID, err)
			return false
		}
	}
	return true
}

// 添加一个成员
func (team *Team) TeamAddUser(user *User) bool {
	// 建立关系
	tu := &TeamUser{
		TeamID: team.ID,
		UserID: user.ID,
	}
	if err := db.Db.Create(tu).Error; err != nil {
		helpers.Log.Error("团队【%d】增加成员【%d】失败:%v", team.ID, user.ID, err)
		return false
	}
	// 增加团队成员数量
	team.IncreaseMemberCount(1)
	// 更新用户的所属团队
	user.BelongTeam = team
	// 如果有申请，则改变状态
	return true
}

// 查找团队的成员列表
func (team *Team) TeamMembers() {

}

// 查找团队的成员数量、本周数据；所辖团队数量，所辖团队本周数据
func (team *Team) GetTeamInfo() {
	// 1. 查询团队成员列表
	teamUsers := make([]*TeamUser, 0)
	db.Db.Where("team_id=?", team.ID).Preload("RUser").Find(&teamUsers)
	team.TeamUsers = teamUsers
	teamUserIds := make([]uint, 0)
	for i := 0; i < len(teamUsers); i++ {
		teamUserIds = append(teamUserIds, teamUsers[i].UserID)
	}
	// 查询团队成员的本周数据
	// 本周周报的日期
	days := helpers.GetPerDayTimeInWeek(time.Now())
	weekReportKey := helpers.GetDayReportKey(days[6])
	reports := make([]*Report, 0)
	db.Db.Where("user_id IN (?) AND `key`=? AND report_type=?", teamUserIds, weekReportKey, WeekReportType).Preload("ReportDatas").Find(&reports)
	// 将数据按照用户id处理成map
	reportUserMap := make(map[uint]*Report)
	for i := 0; i < len(reports); i++ {
		reportUserMap[reports[i].UserID] = reports[i]
	}
	for i := 0; i < len(teamUsers); i++ {
		if _, ok := reportUserMap[teamUsers[i].UserID]; ok {
			teamUsers[i].RUser.WeekReport = reportUserMap[teamUsers[i].UserID]
		} else {
			teamUsers[i].RUser.WeekReport = &Report{}
		}
	}
}

// 查询所辖团队
func (team *Team) GetChildTeams() {
	teams := make([]*Team, 0)
	db.Db.Where("parent_id=?", team.ID).Find(&teams)
	// 查询每个团队的信息
	for i := 0; i < len(teams); i++ {
		teams[i].GetTeamInfo()
	}
	team.Teams = teams
}
