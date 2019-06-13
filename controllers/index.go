package controllers

import (
	"encoding/xml"
	"fmt"
	"github.com/satori/go.uuid"
	"supermentor/helpers"
	"supermentor/models"
)

type AppUpdateStruct struct {
	XMLName xml.Name `xml:"update"`
	Version int64    `xml:"version"`
	Name    string   `xml:"name"`
	Url     string   `xml:"url"`
}

// 检查版本更新
func CheckAndroidUpdateAction(route *SmRoute) {
	appUpdate := AppUpdateStruct{Version: 2, Name: "0.0.2版本", Url: "http://aaa"}
	b, _ := xml.Marshal(appUpdate)
	fmt.Fprintf(route.Response, "%s", string(b))
}

func GetTip(route *SmRoute) {
	tip := models.GetTips()
	route.ReturnJson(CODE_OK, "", tip)
}

// 查询所有能力值
func GetAbilities(route *SmRoute) {
	abilities := models.GetAllAbility()
	route.ReturnJson(CODE_OK, "", abilities)
	return
}

func GetAllCompanies(route *SmRoute) {
	companies := models.GetAllCompanies()
	route.ReturnJson(CODE_OK, "", companies)
	return
}

// 生成一个随机的设备ID
func GetDeviceId(route *SmRoute) {
	u1, err := uuid.NewV4()
	if err != nil {
		route.ReturnJson(CODE_ERROR, "无法为设备生成id", "")
		return
	}
	res := make(map[string]interface{})
	res["device_id"] = u1
	route.ReturnJson(CODE_OK, "", res)
	return
}

// 生成上传凭证，默认生成静态文件的uptoken
func GetUpToken(route *SmRoute) {
	upType := 1
	typeString := route.Request.Form.Get("type")
	if typeString == "2" {
		upType = 2
	}
	if typeString == "3" {
		upType = 3
	}
	token := ""
	qn := helpers.GetQiniu()
	if upType == 1 {
		token = qn.GetStaticFileUpToken()
	}
	if upType == 2 {
		token = qn.GetAudioFileUpToken()
	}
	if upType == 3 {
		token = qn.GetUEditorImageUpToken()
	}
	route.ReturnJson(CODE_OK, "", token)
	return
}

// 发送验证码
func SendCode(route *SmRoute) {
	templateCode := route.Request.Form.Get("t")
	phoneNumber := route.Request.Form.Get("phone")
	if templateCode == "" {
		route.ReturnJson(CODE_ERROR, "发送失败，请重试", "")
		return
	}
	sms := &models.Sms{
		Mobile:       phoneNumber,
		TemplateCode: templateCode,
		Params:       "",
		Data:         make(map[string]string),
	}
	if !sms.SendCode() {
		route.ReturnJson(CODE_ERROR, "发送失败，请重试", "")
		return
	}
	route.ReturnJson(CODE_OK, "", "")
	return
}

// 输出app配置
func GetAppConfig(route *SmRoute) {
	data := make(map[string]interface{})
	data["showTmpLoginBtn"] = 1 // 显示演示按钮
	route.ReturnJson(CODE_OK, "", data)
	return
}

// 加密字符串
func Encrypt(route *SmRoute) {
	s := route.Request.Form.Get("s")
	encryptIDByte, _ := helpers.DesEncrypt([]byte(s), []byte("sm201807"))
	route.ReturnJson(CODE_OK, "", encryptIDByte)
}
