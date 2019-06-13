package controllers

import (
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"strconv"
	"supermentor/helpers"
	"supermentor/models"
	"time"
)

func AdminLogin(route *SmRoute) {
	username := "admin"
	password := "supermentor"

	helpers.Log.Info("%v", route.Request)
	clientUserName := route.Request.Form.Get("username")
	clientPassword := route.Request.Form.Get("password")
	helpers.Log.Info("%s-%s", clientUserName, clientPassword)
	if clientPassword == password && clientUserName == username {
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"user_id": 1,
			"iss":     "sm",
			"sub":     "web",
			"ext":     time.Now().Add(time.Hour * time.Duration(543120)).Unix(),
			"iat":     time.Now().Unix(),
		})
		tokenString, err := token.SignedString([]byte("supermentor-2018-v1.0"))
		if err != nil {
			route.ReturnJson(CODE_ERROR, err.Error(), "")
			return
		}
		// 通过
		res := make(map[string]interface{})
		res["token"] = tokenString
		route.ReturnJson(CODE_OK, "", res)
		return
	}

	route.ReturnJson(CODE_ERROR, "用户名或密码错误", "")
}

// 查询课程列表
func AdminCourseList(route *SmRoute) {
	page := 1
	pageString := route.Request.Form.Get("page")
	if pageString != "" {
		pageInt, pageErr := strconv.Atoi(pageString)
		if pageErr != nil {
			page = pageInt
		}
	}
	courses, total := models.GetCourseList(page)
	res := make(map[string]interface{})
	res["data"] = courses
	res["total"] = total
	route.ReturnJson(CODE_OK, "", res)
}

// 添加课程
func AdminAddCourse(route *SmRoute) {
	courseId := 0
	idString := route.Request.Form.Get("id")
	if idString != "" {
		id, err := strconv.Atoi(idString)
		if err == nil && id > 0 {
			courseId = id
		}
	}
	helpers.Log.Info("idString: %s, id: %d", idString, courseId)
	title := route.Request.Form.Get("title")
	cover := route.Request.Form.Get("cover")
	desc := route.Request.Form.Get("desc")
	content := route.Request.Form.Get("content")
	audio := route.Request.Form.Get("audio")
	abilityIdString := route.Request.Form.Get("abilityId")
	abilityId := 0
	if abilityIdString != "" {
		abilityIdInt, aErr := strconv.Atoi(abilityIdString)
		if aErr != nil {
			route.ReturnJson(CODE_ERROR, "请选择一个能立项", "")
			return
		}
		abilityId = abilityIdInt
	}
	durationString := route.Request.Form.Get("duration")
	duration := 0.0
	if durationString != "" {
		durationFloat, dErr := strconv.ParseFloat(durationString, 64)
		if dErr != nil {
			route.ReturnJson(CODE_ERROR, "音频时长不能为0", "")
			return
		}
		duration = durationFloat
	}
	fmt.Println(title, cover, desc, content, audio, duration)
	if title == "" || cover == "" || desc == "" || content == "" {
		route.ReturnJson(CODE_ERROR, "请填写完整的课程信息", "")
		return
	}
	// 检查标题是否重复
	course := models.GetCourseById(uint(courseId))
	course.Title = title
	course.Cover = cover
	course.AbilityID = uint(abilityId)
	course.Desc = desc
	course.Content = content
	course.AudioDuration = duration

	err := course.Save()
	if err != nil {
		route.ReturnJson(CODE_ERROR, err.Error(), "")
		return
	}
	route.ReturnJson(CODE_OK, "", "")
}
