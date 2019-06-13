package models

import (
	"encoding/json"
	"fmt"
	"github.com/jinzhu/gorm"
	"strconv"
	"supermentor/db"
	"supermentor/helpers"
	"time"
)

// 签名
const SmsSignName = "超级师父"

// 超时时间
const SmsDuration = 900

type SmsStatus int

const (
	_ SmsStatus = iota
	SmsSending
	SmsSendFailed
	SmsSendOK
)

type Sms struct {
	gorm.Model
	Mobile       string
	TemplateCode string
	Params       string
	Status       SmsStatus
	Duration     int
	BizID        string

	Data map[string]string `gorm:"-"`
}

func (Sms) TableName() string {
	return "sms"
}

// 发送验证码，如果上一条发送时间在一分钟内，则不允许发送
// 如果之前的验证码没有使用过，则继续发送老的验证码
//
func (sms *Sms) SendCode() bool {
	if !sms.checkCanSend() {
		// 不能发送
		helpers.Log.Warning("%s一分钟内只能发送一次%s短信", sms.Mobile, sms.TemplateCode)
		return false
	}
	// 查找老的码
	code := sms.GetSmsCode()
	// 拼接参数
	sms.Data["code"] = code
	paramsJson, _ := json.Marshal(sms.Data)
	paramsString := string(paramsJson[:])
	sms.Params = paramsString
	sms.Status = SmsSending
	sms.Duration = SmsDuration
	// 创建一条记录
	if err := db.Db.Create(sms).Error; err != nil {
		helpers.Log.Warning("%s发送%s短信入库失败：%+v", sms.Mobile, sms.TemplateCode, err)
		return false
	}
	if sms.SendSms() {
		sms.Status = SmsSendOK
		db.Db.Model(sms).Update("status", SmsSendOK)
	}
	return true
}

// 检查一分钟内是否发送过
func (sms *Sms) checkCanSend() bool {
	oldSms := Sms{}
	if db.Db.Where("mobile=? AND template_code=? AND status=?", sms.Mobile, sms.TemplateCode, SmsSendOK).Order("id DESC").Find(&oldSms).RecordNotFound() {
		return true
	}
	// 检查时间
	now := time.Now()
	subM := now.Sub(oldSms.CreatedAt).Minutes()
	// 1分钟内发送过，则不能发送
	if subM < 1 {
		return false
	}
	return true
}

// 发验短信
func (sms *Sms) SendSms() bool {
	client := helpers.GetAliyunClient()
	params := make(map[string]string)
	response := make(map[string]string)
	params["Version"] = "2017-05-25"
	params["RegionId"] = "cn-hangzhou"
	params["PhoneNumbers"] = sms.Mobile
	params["SignName"] = SmsSignName
	params["TemplateParam"] = sms.Params
	params["TemplateCode"] = sms.TemplateCode
	cerr := client.DoAliyunAction("SendSms", params, &response)
	if cerr != nil {
		helpers.Log.Warning("短信发送失败：%v", cerr)
		return false
	}
	// 判断是否成功
	if code, exists := response["Code"]; !exists || code != "OK" {
		// 发送失败
		sms.Status = SmsSendFailed
		// 更新
		db.Db.Model(sms).Update("status", SmsSendFailed)
		return false
	}
	sms.BizID = response["BizId"]
	sms.Status = SmsSendOK
	db.Db.Model(sms).Update(sms)
	return true
}

// 生成一条验证码，存储在缓存中
func (sms *Sms) GetSmsCode() string {
	var code string
	if code = sms.GetCodeFromCache(); code == "" {
		// 已经有了，发送旧的
		code = helpers.RandNumber(4)
		// 入缓存
		sms.SetCodeToCache(code)
	}
	return code
}

// 从缓存中查询验证码是否存在
func (sms *Sms) GetCodeFromCache() string {
	var code string
	cacheKey := sms.GetSmsCacheKey()
	code, err := db.Redis.Get(cacheKey).Result()
	if err != nil {
		helpers.Log.Info("get cached code failed：%v", err)
		return ""
	}
	helpers.Log.Info("cached code：%s => %s", cacheKey, code)
	return code
}

// 将验证码存入缓存
func (sms *Sms) SetCodeToCache(code string) bool {
	cacheKey := sms.GetSmsCacheKey()
	duration := fmt.Sprintf("%ds", sms.Duration)
	expiry, _ := time.ParseDuration(duration)
	if err := db.Redis.Set(cacheKey, code, expiry).Err(); err != nil {
		helpers.Log.Error("redis 写入失败【%s】=>【%s】: %v", cacheKey, code, err)
		return false
	}
	return true
}

// 获取短信缓存key
func (sms *Sms) GetSmsCacheKey() string {
	return "sms-" + sms.TemplateCode + "-" + sms.Mobile
}

// 检查验证码
func CheckSmsCode(mobile string, templateCode string, code string) bool {
	sms := &Sms{
		Mobile:       mobile,
		TemplateCode: templateCode,
	}
	key := sms.GetSmsCacheKey()
	cacheCode := sms.GetCodeFromCache()
	if cacheCode == "" {
		return false
	}
	if code == cacheCode {
		// 删除code
		db.Redis.Del(key)
		db.Redis.Del(key + "-used")
		return true
	}
	c := db.Redis.Get(key + "-used").String()
	count, err := strconv.Atoi(c)
	duration := fmt.Sprintf("%ds", SmsDuration)
	expiry, _ := time.ParseDuration(duration)
	if err != nil || count == 0 || c == "" {
		count = 1
	} else {
		count++
	}
	db.Redis.Set(key+"-used", count, expiry)
	return false
}
