package controllers

import (
	"encoding/json"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"net/http"
	"strings"
	"supermentor/helpers"
)

type ResCode int

const (
	_              ResCode = 1
	CODE_OK                = 200
	CODE_NOT_LOGIN         = 401
	CODE_ERROR             = 500
)

type SmHandler interface {
	ServeHTTP(http.ResponseWriter, *http.Request)
	CheckLogin()
	ReturnJson(ResCode, string, interface{})
}

type SmRoute struct {
	Pattern   string              // 路径
	IsAdmin   bool                // 是否后台接口
	NeedLogin bool                // 是否需要登录
	UserId    uint                // 当前登录用户，默认为0
	Response  http.ResponseWriter // response对象
	Request   *http.Request       // request对象
	Handler   func(*SmRoute)      // 处理方法
}

type ResponseObject struct {
	Code ResCode     `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

type EmptyData struct{}

type SmClaims struct {
	UserID uint `json:"user_id"`
	jwt.StandardClaims
}

func (route *SmRoute) ReturnJson(code ResCode, msg string, data interface{}) {
	if data == "" {
		data = EmptyData{}
	}
	resObj := ResponseObject{
		Code: code,
		Msg:  msg,
		Data: data,
	}
	helpers.Log.Info("%+v", resObj)
	resBytes, _ := json.Marshal(resObj)
	route.Response.Write(resBytes)
}

func (route SmRoute) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	route.Response = w
	route.Request = r
	err := route.Request.ParseForm()
	if err != nil {
		fmt.Printf("parse form failed: %v\n", err)
	}
	w.Header().Add("Access-Control-Allow-Origin", "*")
	w.Header().Add("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, HEAD, OPTIONS")
	w.Header().Add("Access-Control-Allow-Headers", "did, Authorization, Content-Type")
	w.Header().Set("Content-Type", "application/json;charset=utf-8")
	var uid uint = 0
	if r.Method != "POST" && r.Method != "GET" {
		route.ReturnJson(CODE_OK, "", "")
		return
	}
	if route.NeedLogin {
		// 检查登录
		uid = route.CheckLogin()
		if uid == 0 {
			route.ReturnJson(CODE_NOT_LOGIN, "user not login", "")
			return
		}
	}
	route.UserId = uid
	route.Handler(&route)
	return
}

// 返回登录用户id
func (route *SmRoute) CheckLogin() uint {
	tokenString := route.Request.Header.Get("Authorization")
	if tokenString == "" {
		return 0
	}
	tokenString = strings.Replace(tokenString, "Bearer ", "", 1)
	token, err := jwt.ParseWithClaims(tokenString, &SmClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte("supermentor-2018-v1.0"), nil
	})
	if err != nil || !token.Valid {
		return 0
	}
	claims := token.Claims.(*SmClaims)
	return claims.UserID
}
