package helpers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

var WechatAppId string = ""
var WechatAppSecret string = ""

var WechatMpAppId string = ""
var WechatMpAppSecret string = ""

type WechatAuthStruct struct {
	AccessToken string `json:"access_token"`
	OpenId      string `json:"openid"`
}

type WechatUserStruct struct {
	NickName string `json:"nickname"`
	OpenId   string `json:"openid"`
	UniId    string `json:"unionid"`
	Sex      uint   `json:"sex"`
	Avatar   string `json:"headimgurl"`
	City     string `json:"city"`
	Province string `json:"province"`
	Country  string `json:"country"`
}

func GetAccessToken(appId string, appSecret string, code string) (string, string, error) {
	fmt.Printf("%s, %s", WechatAppId, WechatAppSecret)
	// 请求微信，然后获取accesstoken和openid
	url := fmt.Sprintf("https://api.weixin.qq.com/sns/oauth2/access_token?appid=%s&secret=%s&code=%s&grant_type=authorization_code", appId, appSecret, code)
	Log.Info("request wechat: %s", url)
	resp, err := http.Get(url)
	if err != nil {
		Log.Error("wechat: %v", err)
		return "", "", err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		// handle error
		Log.Error("wechat: %v", err)
		return "", "", err
	}
	Log.Info("wechat response: %s", body)
	wechatStruct := &WechatAuthStruct{}
	_ = json.Unmarshal(body, wechatStruct)
	Log.Info("wechat response to JSON: %+v", wechatStruct)
	accessToken := wechatStruct.AccessToken
	openId := wechatStruct.OpenId
	return accessToken, openId, nil
}

func GetUserInfo(accessToken string, openId string) (WechatUserStruct, error) {
	userInfoBytes := WechatUserStruct{}
	url := fmt.Sprintf("https://api.weixin.qq.com/sns/userinfo?access_token=%s&openid=%s&lang=zh_CN", accessToken, openId)
	Log.Info("request wechat userInfo: %s", url)
	resp, err := http.Get(url)
	if err != nil {
		// handle error
		Log.Error("wechat: %v", err)
		return userInfoBytes, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		// handle error
		Log.Error("wechat: %v", err)
		return userInfoBytes, err
	}
	Log.Info("wechat userInfo Response: %s", body)
	json.Unmarshal(body, &userInfoBytes)
	Log.Info("wechat userInfo JSON: %+v", userInfoBytes)
	return userInfoBytes, nil
}
