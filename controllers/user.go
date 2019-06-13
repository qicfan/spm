package controllers

import (
	"github.com/dgrijalva/jwt-go"
	"strconv"
	"supermentor/helpers"
	"supermentor/models"
	"time"
)

// 登录
// 检查是否已经登录
// 1. 跳转到微信授权
// 2. 微信回跳，获取用户基本信息
// 3. 检查用户是否已经注册
// 4. 如果已经注册，则写入 cookie，然后跳转到登录前页面，如果没有登录前页面，则跳转到首页
func LoginAction(route *SmRoute) {
	user := &models.User{}
	r := route.Request
	code := r.Form.Get("code") // 得到querystring的forward参数
	if code == "" {
		route.ReturnJson(CODE_ERROR, "参数错误", "")
		return
	}
	// 获取微信授权
	accessToken, openId, tokenErr := helpers.GetAccessToken(helpers.WechatAppId, helpers.WechatAppSecret, code)
	if tokenErr != nil {
		route.ReturnJson(CODE_ERROR, "wechat return error", "")
		return
	}
	wechatUser, wechatErr := helpers.GetUserInfo(accessToken, openId)
	if wechatErr != nil {
		route.ReturnJson(CODE_ERROR, "wechat return error", "")
		return
	}
	// 查询用户是否存在
	user, userErr := models.GetUserByPlatformId("wechat", wechatUser.OpenId)
	if userErr != nil || user.ID == 0 {
		// 创建一个新用户
		createErr := user.CreateWechatUser(&wechatUser)
		if createErr != nil {
			route.ReturnJson(CODE_ERROR, createErr.Error(), "")
			return
		}
	}
	models.RecordUserAction(user.ID, models.UserLoginAction, "微信登录")
	user.Password = "" // 不能返回密码
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"iss":     "sm",
		"sub":     "",
		"ext":     time.Now().Add(time.Hour * time.Duration(543120)).Unix(),
		"iat":     time.Now().Unix(),
	})
	tokenString, err := token.SignedString([]byte("supermentor-2018-v1.0"))
	if err != nil {
		route.ReturnJson(CODE_ERROR, err.Error(), "")
		return
	}
	res := make(map[string]interface{})
	res["user"] = user
	res["is_new"] = user.IsNew
	res["token"] = tokenString
	route.ReturnJson(CODE_OK, "", res)
}

// 查询用户信息
func GetUserInfo(route *SmRoute) {
	user, err := models.GetUserById(route.UserId, true)
	if err != nil {
		route.ReturnJson(CODE_ERROR, "user not found", "")
		return
	}
	route.ReturnJson(CODE_OK, "", user)
}

// 查询推荐给用户的课程
func GetUserRecommendCourse(route *SmRoute) {
	isFinishString := route.Request.Form.Get("is_finish")
	if isFinishString == "" {
		isFinishString = "0"
	}
	isFinish, isFinishErr := strconv.Atoi(isFinishString)
	if isFinishErr != nil {
		isFinish = 0
	}
	courses := models.GetUserRecommendCourse(route.UserId, isFinish)
	route.ReturnJson(CODE_OK, "", map[string]interface{}{
		"data": courses,
	})
}

// 查询用户一段时间内的行为总数
func GetUserActionCount(route *SmRoute) {
	startTimeString := route.Request.Form.Get("start_time")
	endTimeString := route.Request.Form.Get("end_time")
	if startTimeString == "" || endTimeString == "" {
		route.ReturnJson(CODE_ERROR, "参数错误", "")
		return
	}
	startTime, _ := strconv.ParseInt(startTimeString, 10, 64)
	endTime, _ := strconv.ParseInt(endTimeString, 10, 64)
	actions := models.GetUserActionCount(route.UserId, startTime, endTime)
	route.ReturnJson(CODE_OK, "", actions)
}

// 查询用户的报告
func GetUserReport(route *SmRoute) {
	page := route.Request.Form.Get("p")
	p, err := strconv.Atoi(page)
	if err != nil {
		p = 1
	}
	reports := models.GetUserReport(route.UserId, p)
	route.ReturnJson(CODE_OK, "", reports)
}

// 查询用户的课程
func GetUserCourseById(route *SmRoute) {
	courseIdString := route.Request.Form.Get("id")
	if courseIdString == "" {
		route.ReturnJson(CODE_ERROR, "参数错误", "")
		return
	}
	courseId, _ := strconv.Atoi(courseIdString)
	userCourse := models.GetUserCourseByID(uint(courseId), route.UserId)
	route.ReturnJson(CODE_OK, "", userCourse)
}

// 更新昵称
func UpdateUserNickName(route *SmRoute) {
	user := &models.User{}
	newName := route.Request.Form.Get("nickname")
	if newName == "" {
		route.ReturnJson(CODE_OK, "", user)
		return
	}
	user, userErr := models.GetUserById(route.UserId, false)
	if userErr != nil {
		route.ReturnJson(CODE_OK, "", user)
		return
	}
	user.UpdateName(newName)
	route.ReturnJson(CODE_OK, "", user)
	return
}

// 更新用户公司
func UpdateUserCompany(route *SmRoute) {
	user := &models.User{}
	companyIdString := route.Request.Form.Get("company_id")
	if companyIdString == "" {
		route.ReturnJson(CODE_OK, "", user)
		return
	}
	companyId, err := strconv.Atoi(companyIdString)
	if err != nil {
		route.ReturnJson(CODE_OK, "", user)
		return
	}
	company := models.GetCompanyById(uint(companyId))
	if company.ID == 0 {
		route.ReturnJson(CODE_OK, "", user)
		return
	}
	user, userErr := models.GetUserById(route.UserId, false)
	if userErr != nil {
		route.ReturnJson(CODE_OK, "", user)
		return
	}
	user.UpdateCompany(company)
	route.ReturnJson(CODE_OK, "", user)
	return
}

// 绑定手机号
func UpdateUserMobile(route *SmRoute) {
	user := &models.User{}
	mobile := route.Request.Form.Get("mobile")
	code := route.Request.Form.Get("code")
	templateCode := route.Request.Form.Get("t")
	if mobile == "" || code == "" || templateCode == "" {
		route.ReturnJson(CODE_OK, "", user)
		return
	}
	if !models.CheckSmsCode(mobile, templateCode, code) {
		route.ReturnJson(CODE_ERROR, "验证码错误", user)
		return
	}
	user, userErr := models.GetUserById(route.UserId, false)
	if userErr != nil {
		route.ReturnJson(CODE_ERROR, "用户不存在", user)
		return
	}
	// 查看手机号是否被使用
	newUser := models.GetUserByMobile(mobile, true)
	if newUser.ID > 0 {
		route.ReturnJson(CODE_ERROR, "该手机号已经被使用", user)
		return
	}
	user.UpdateMobile(mobile)
	route.ReturnJson(CODE_OK, "", user)
	return
}

// 用设备id进行登录
func TmpAuth(route *SmRoute) {
	did := route.Request.Form.Get("did")
	if did == "" {
		route.ReturnJson(CODE_ERROR, "参数错误", "")
		return
	}
	// 查询用户，没有的话就创建
	user, userErr := models.GetUserByPlatformId("device", did)
	isNew := 0
	if userErr != nil || user.ID == 0 {
		// 创建一个新用户
		createErr := user.CreateDeviceUser(did)
		if createErr != nil {
			route.ReturnJson(CODE_ERROR, createErr.Error(), "")
			return
		}
		isNew = 1
	}
	models.RecordUserAction(user.ID, models.UserLoginAction, "临时登录")
	user.Password = "" // 不能返回密码
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"iss":     "sm",
		"sub":     did,
		"ext":     time.Now().Add(time.Hour * time.Duration(543120)).Unix(),
		"iat":     time.Now().Unix(),
	})
	tokenString, err := token.SignedString([]byte("supermentor-2018-v1.0"))
	if err != nil {
		route.ReturnJson(CODE_ERROR, err.Error(), "")
		return
	}
	res := make(map[string]interface{})
	res["user"] = user
	res["is_new"] = isNew
	res["token"] = tokenString
	route.ReturnJson(CODE_OK, "", res)
}

func UserFinishGuide(route *SmRoute) {
	user, _ := models.GetUserById(route.UserId, false)
	user.FinishGuide()
	route.ReturnJson(CODE_OK, "", "")
}

func UserFinishEvaluating(route *SmRoute) {
	user, _ := models.GetUserById(route.UserId, false)
	user.FinishEvaluating()
	route.ReturnJson(CODE_OK, "", "")
}
