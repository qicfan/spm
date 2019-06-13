package helpers

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/satori/go.uuid"
	"io/ioutil"
	"net/http"
	"sort"
	"strings"
	"time"
)

// 阿里云客户端
type AliyunClient struct {
	AccessKey string
	SecretKey string
	Host      string
	Params    map[string]string
}

// 阿里云短信对象
type AliyunSmsParams struct {
	Action        string
	Version       string
	RegionId      string
	PhoneNumbers  string
	SignName      string
	TemplateCode  string
	TemplateParam string
}

var AliyunAK string
var AliyunSK string

// 初始化系统参数
func (client *AliyunClient) GetAliyunSystemPrams() {
	u1, _ := uuid.NewV4()
	curtime := time.Now()
	hh, _ := time.ParseDuration("-8h")
	now := curtime.Add(hh)
	ts := now.Format("2006-01-02T15:04:05Z")
	client.Params["SignatureMethod"] = "HMAC-SHA1"
	client.Params["SignatureNonce"] = hex.EncodeToString(u1[:])
	client.Params["AccessKeyId"] = client.AccessKey
	client.Params["SignatureVersion"] = "1.0"
	client.Params["Timestamp"] = ts
	client.Params["Format"] = "JSON"
}

// 创建一个阿里云客户端
func GetAliyunClient() *AliyunClient {
	client := &AliyunClient{
		AccessKey: AliyunAK,
		SecretKey: AliyunSK,
		Host:      "http://dysmsapi.aliyuncs.com/",
		Params:    make(map[string]string),
	}
	return client
}

// 执行一个请求
func (client *AliyunClient) DoAliyunAction(action string, params map[string]string, response interface{}) error {
	client.GetAliyunSystemPrams()
	// 1. 合并参数，params会覆盖clien.Params里的同名参数
	for k, v := range params {
		client.Params[k] = v
	}
	client.Params["Action"] = action
	// 删除掉签名参数
	if _, exists := client.Params["Signature"]; exists {
		delete(client.Params, "Signature")
	}
	Log.Info("aliyun params: %+v", client.Params)
	// 2. 根据key的ascii从小到大排列
	keys := make([]string, 0)
	for k := range client.Params {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	Log.Info("aliyun sort keys: %v", keys)
	// 3. 拼接参数
	columns := make([]string, 0)
	for i := 0; i < len(keys); i++ {
		k := keys[i]
		v := client.Params[k]
		p := SpecialUrlEncode(k) + "=" + SpecialUrlEncode(v)
		columns = append(columns, p)
	}
	queryString := strings.Join(columns, "&")
	Log.Info("aliyun columns: %s", queryString)
	// 4. 拼接请求参数
	SignString := fmt.Sprintf("GET&%s&%s", SpecialUrlEncode("/"), SpecialUrlEncode(queryString))
	Log.Info("aliyun queryString no sign: %s", SignString)
	// 5. 签名
	sign := AliyunSign(client.SecretKey+"&", SignString)
	Log.Info("aliyun sign: %s", sign)
	// 6. 拼接签名到请求参数中
	queryString = fmt.Sprintf("Signature=%s&%s", SpecialUrlEncode(sign), queryString)
	Log.Info("aliyun queryString has sign: %s", queryString)
	// 7. 发送请求
	url := fmt.Sprintf("%s?%s", client.Host, queryString)
	Log.Info("request aliyun: %s", url)
	resp, err := http.Get(url)
	if err != nil {
		// handle error
		Log.Error("aliyun: %v", err)
		return err
	}
	defer resp.Body.Close()
	// 8. 解析返回值
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		// handle error
		Log.Error("aliyun: %v", err)
		return err
	}
	Log.Info("aliyun Response: %s", body)
	json.Unmarshal(body, &response)
	Log.Info("aliyun response JSON: %+v", response)
	return nil
}

// 生成签名
func AliyunSign(secretKey string, stringToSign string) string {
	mac := hmac.New(sha1.New, []byte(secretKey))
	mac.Write([]byte(stringToSign))
	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}
