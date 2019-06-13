package controllers

import (
	"strconv"
	"supermentor/helpers"
	"supermentor/models"
)

// 创建团队
func CreateTeam(route *SmRoute) {
	userId := route.UserId
	// 查找用户
	user, _ := models.GetUserById(userId, true)
	if user.ID == 0 {
		route.ReturnJson(CODE_OK, "", "")
		return
	}
	if user.MyTeam.ID > 0 {
		// 已经有团队了
		route.ReturnJson(CODE_OK, "", user)
		return
	}
	// 判断是否可以创建
	if user.CompanyId == 0 || user.Mobile == "" || user.Nickname == "" {
		// 不符合创建规则
		route.ReturnJson(CODE_OK, "", user)
		return
	}
	models.CreateTeam(user)
	route.ReturnJson(CODE_OK, "", user)
	return
}

func SearchTeamByMobile(route *SmRoute) {
	mobile := route.Request.Form.Get("mobile")
	searchType := route.Request.Form.Get("type")
	if mobile == "" || searchType == "" {
		route.ReturnJson(CODE_OK, "", "")
		return
	}
	user := models.GetUserByMobile(mobile, true)
	hasTeam := false
	member := &models.User{}
	leader := &models.User{}
	if searchType == "leader" {
		// 查找团队长
		if user.MyTeam.ID > 0 {
			// 有所属团队，是团队长
			hasTeam = true
			leader = user
		}
	} else {
		// 查找团队成员
		if user.BelongTeam.ID > 0 {
			// 有所属团队
			hasTeam = true
			member = user
			// 查找所属团队的owner
			leader, _ = models.GetUserById(user.BelongTeam.OwnerID, true)
		}
	}

	res := make(map[string]interface{})
	res["has-team"] = hasTeam
	res["leader"] = leader
	res["member"] = member

	route.ReturnJson(CODE_OK, "", res)
	return
}

// 离开团队
func LeaveTeam(route *SmRoute) {
	userId := route.UserId
	user, _ := models.GetUserById(userId, true)
	if user.ID == 0 {
		// 没有团队
		route.ReturnJson(CODE_OK, "", "")
	}
	rs := user.LeaveTeam()
	route.ReturnJson(CODE_OK, "", rs)
}

// 移除组员
func DeleteUserFromTeam(route *SmRoute) {
	userIdString := route.Request.Form.Get("user_id")
	if userIdString == "" {
		route.ReturnJson(CODE_ERROR, "参数错误", "")
		return
	}
	userId, err := strconv.Atoi(userIdString)
	if err != nil {
		route.ReturnJson(CODE_ERROR, "参数错误", "")
		return
	}
	user, _ := models.GetUserById(uint(userId), true)
	if user.ID == 0 {
		// 没有团队
		route.ReturnJson(CODE_OK, "", "")
	}
	rs := user.LeaveTeam()
	route.ReturnJson(CODE_OK, "", rs)
}

// 发送一个加入团队的申请
func ApplyTeam(route *SmRoute) {
	teamOwnerIdString := route.Request.Form.Get("team_owner_id")
	if teamOwnerIdString == "" {
		route.ReturnJson(CODE_ERROR, "请选择一个团队加入", "")
		return
	}
	// 尝试进行解密
	decodeTeamOwnerIDString, desErr := helpers.DesDecrypt(teamOwnerIdString, []byte("sm201807"))
	if desErr == nil && decodeTeamOwnerIDString != "" {
		// 成功解密，使用
		teamOwnerIdString = decodeTeamOwnerIDString
	}
	teamOwnerId, teamOwnerIdErr := strconv.Atoi(teamOwnerIdString)
	if teamOwnerIdErr != nil {
		route.ReturnJson(CODE_ERROR, "请选择一个团队加入", "")
		return
	}
	userId := route.UserId
	// 查找用户
	user, _ := models.GetUserById(userId, true)
	// 查找团队长
	teamOwner, _ := models.GetUserById(uint(teamOwnerId), true)
	if user.BelongTeam.ID > 0 {
		route.ReturnJson(CODE_ERROR, "请先退出当前团队再加入其它团队", "")
		return
	}
	if teamOwner.MyTeam.ID == 0 {
		// 如果团队长没有团队，则创建一个
		//route.ReturnJson(CODE_ERROR, "所选团队不存在", "")
		//return
		models.CreateTeam(teamOwner)

	}
	// 添加团队申请
	rs := teamOwner.MyTeam.TeamAddApply(user)
	if !rs {
		route.ReturnJson(CODE_ERROR, "申请加入团队失败，请稍后再试", "")
		return
	}
	route.ReturnJson(CODE_OK, "", "")
}

// 驳回加入申请
func TeamApplyReject(route *SmRoute) {
	userIdString := route.Request.Form.Get("user_id")
	if userIdString == "" {
		route.ReturnJson(CODE_ERROR, "该团队加入申请不存在", "")
		return
	}
	userId, teamOwnerIdErr := strconv.Atoi(userIdString)
	if teamOwnerIdErr != nil {
		route.ReturnJson(CODE_ERROR, "该团队加入申请不存在", "")
		return
	}
	teamOwnerId := route.UserId
	// 查找用户
	user, _ := models.GetUserById(uint(userId), true)
	// 查找团队长
	teamOwner, _ := models.GetUserById(teamOwnerId, true)
	if user.BelongTeam.ID > 0 {
		route.ReturnJson(CODE_ERROR, "用户已加入其它团队", "")
		return
	}
	if teamOwner.MyTeam.ID == 0 {
		route.ReturnJson(CODE_ERROR, "您还没有创建团队，无法加入成员", "")
		return
	}
	ta := models.GetTeamApply(user, teamOwner.MyTeam)
	if ta.ID == 0 {
		route.ReturnJson(CODE_ERROR, "该团队加入申请不存在", "")
		return
	}
	ta.TeamApplyReject()
	route.ReturnJson(CODE_OK, "", "")
}

// 批准加入团队
func TeamApplyPass(route *SmRoute) {
	userIdString := route.Request.Form.Get("user_id")
	if userIdString == "" {
		route.ReturnJson(CODE_ERROR, "该团队加入申请不存在", "")
		return
	}
	userId, teamOwnerIdErr := strconv.Atoi(userIdString)
	if teamOwnerIdErr != nil {
		route.ReturnJson(CODE_ERROR, "该团队加入申请不存在", "")
		return
	}
	teamOwnerId := route.UserId
	// 查找用户
	user, _ := models.GetUserById(uint(userId), true)
	// 查找团队长
	teamOwner, _ := models.GetUserById(teamOwnerId, true)
	if user.BelongTeam.ID > 0 {
		route.ReturnJson(CODE_ERROR, "用户已加入其它团队", "")
		return
	}
	if teamOwner.MyTeam.ID == 0 {
		route.ReturnJson(CODE_ERROR, "您还没有创建团队，无法加入成员", "")
		return
	}
	ta := models.GetTeamApply(user, teamOwner.MyTeam)
	if ta.ID == 0 {
		route.ReturnJson(CODE_ERROR, "该团队加入申请不存在", "")
		return
	}
	rs := teamOwner.MyTeam.TeamAddUser(user)
	if !rs {
		route.ReturnJson(CODE_ERROR, "审核通过失败，请稍后再试", "")
		return
	}
	ta.TeamApplyPass()
	route.ReturnJson(CODE_OK, "", "")
}

// 查找团队成员，附带每个成员的本周数据
func TeamMembers(route *SmRoute) {
	userId := route.UserId
	// 查找用户
	user, _ := models.GetUserById(userId, true)
	if user.ID == 0 || user.MyTeam.ID == 0 {
		route.ReturnJson(CODE_OK, "", "")
		return
	}
	// 查询团队成员和信息
	user.MyTeam.GetTeamInfo()
	route.ReturnJson(CODE_OK, "", user.MyTeam)
	return
}

// 查询团队信息
func GetTeamInfo(route *SmRoute) {
	userId := route.UserId
	// 查找用户
	user, _ := models.GetUserById(userId, true)
	if user.ID == 0 || user.MyTeam.ID == 0 {
		route.ReturnJson(CODE_OK, "", user.MyTeam)
		return
	}
	// 查询团队成员和信息
	user.MyTeam.GetTeamInfo()
	// 查询所辖团队的信息
	user.MyTeam.GetChildTeams()
	route.ReturnJson(CODE_OK, "", user.MyTeam)
	return
}

// 查找团队加入申请
func TeamApplies(route *SmRoute) {
	pageString := route.Request.Form.Get("page")
	page := 1
	if pageString != "" {
		pageInt, pageErr := strconv.Atoi(pageString)
		if pageErr == nil {
			page = pageInt
		}
	}
	userId := route.UserId
	// 查找用户
	user, _ := models.GetUserById(userId, true)
	if user.ID == 0 {
		route.ReturnJson(CODE_OK, "1", "")
		return
	}
	tas := user.TeamApplies(page)
	route.ReturnJson(CODE_OK, "", tas)
	return
}
