package controllers

import (
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"net/http"
	"net/url"
	"strings"
	"supermentor/helpers"
	"supermentor/models"
	"time"
)

// 组合参数并且跳转到微信平台获取授权和code
func WechatAuth(route *SmRoute) {
	// 前端回调地址
	forward := route.Request.Form.Get("forward")
	if forward == "" {
		route.ReturnJson(500, "参数错误", "")
		return
	}
	host := route.Request.Host
	fmt.Printf("%s\n\n", host)
	fmt.Printf("%s\n\n", forward)
	params := url.Values{}
	params.Set("forward", forward)
	encodeParams := params.Encode()
	redirectUrl := fmt.Sprintf("http://%s/wechat/callback?%s", host, encodeParams)
	redirectParams := url.Values{}
	redirectParams.Set("appid", helpers.WechatMpAppId)
	redirectParams.Set("redirect_uri", redirectUrl)
	redirectParams.Set("response_type", "code")
	redirectParams.Set("scope", "snsapi_userinfo")
	redirectParams.Set("state", "sm")
	wxUrl := fmt.Sprintf("https://open.weixin.qq.com/connect/oauth2/authorize?%s#wechat_redirect", redirectParams.Encode())
	http.Redirect(route.Response, route.Request, wxUrl, http.StatusFound)
	return
}

// 微信授权后携带code参数回调
// 通过code参数获取到openID，通过openID查询用户信息
// 跳转回回调地址
func WechatCallback(route *SmRoute) {
	// 前端回调地址
	forward := route.Request.Form.Get("forward")
	code := route.Request.Form.Get("code")
	fmt.Printf("%s\n\n", code)
	fmt.Printf("%s\n\n", forward)
	if forward == "" || code == "" {
		route.ReturnJson(500, "参数错误", "")
		return
	}
	// 通过code获取OpenID
	// 获取微信授权
	accessToken, openId, tokenErr := helpers.GetAccessToken(helpers.WechatMpAppId, helpers.WechatMpAppSecret, code)
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
	if strings.Contains(forward, "?") {
		forward = fmt.Sprintf("%s&token=%s", forward, tokenString)
	} else {
		forward = fmt.Sprintf("%s?token=%s", forward, tokenString)
	}
	http.Redirect(route.Response, route.Request, forward, http.StatusFound)
	return
}
