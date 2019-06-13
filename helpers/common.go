package helpers

import (
	"crypto/sha1"
	"fmt"
	"github.com/robertkrimen/otto"
	"math/rand"
	"qiniupkg.com/x/url.v7"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// 生成sid
func GetSid(uid string, did string) string {
	signStringBytes := []byte(fmt.Sprintf("%s-%s-sm", did, uid))
	h := sha1.New()
	h.Write(signStringBytes)
	signHex := h.Sum(nil)
	sign := fmt.Sprintf("%x", signHex)
	return sign
}

// 根据给定得时间生成日报和周报key
func GetDayReportKey(date time.Time) string {
	return date.Format("20060102")
}

// 根据给定得时间生成月报key
func GetMonthReportKey(date time.Time) string {
	return date.Format("20060100")
}

// 根据给定得时间生成年报key
func GetYearReportKey(date time.Time) string {
	return date.Format("20060000")
}

// 得到一周里每天得时间
func GetPerDayTimeInWeek(date time.Time) []time.Time {
	days := make([]time.Time, 7)
	todayWeek := int(date.Weekday())
	if todayWeek == 0 {
		// 将周日设置为7
		todayWeek = 7
	}
	// 取得今日零点的时间戳
	todayZero := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	todayZeroUnix := todayZero.Unix()
	// 先处理今日之前的
	for i := 1; i < todayWeek; i++ {
		days[i-1] = time.Unix(todayZeroUnix-int64((todayWeek-i)*86400), 0)
	}
	// 再处理今日之后的
	for i := todayWeek; i <= 7; i++ {
		if i == todayWeek {
			days[i-1] = todayZero
			continue
		}
		days[i-1] = time.Unix(todayZeroUnix+int64((i-todayWeek)*86400), 0)
	}
	return days
}

// 执行表达式，返回计算结果，只能返回一个结果
func ExecJsFormula(formula string, params map[string]float64) float64 {
	vm := otto.New()
	for k, v := range params {
		vm.Set(k, v)
	}
	v, e := vm.Run("var result=" + formula + ";console.log(result);")
	if e != nil {
		// 计算失败，记录日志，返回false
		Log.Error("js虚拟机执行失败：%+v\n\n", e)
		return 0
	}
	fmt.Printf("js虚拟机返回值: %v\n\n", v)
	result, resultE := vm.Get("result")
	if resultE != nil {
		Log.Error("获取js引擎计算结果失败：%v\n\n", e)
		return 0
	}
	fmt.Printf("result = : %v\n\n", result)
	d, toiE := result.ToFloat()
	if toiE != nil {
		Log.Error("js引擎计算结果转Float失败：%v\n\n", e)
		return 0
	}
	fmt.Printf("d = : %v\n\n", d)
	return d
}

// 将struct转成map
func Struct2Map(obj interface{}) map[string]interface{} {
	t := reflect.TypeOf(obj)
	v := reflect.ValueOf(obj)

	var data = make(map[string]interface{})
	for i := 0; i < t.NumField(); i++ {
		data[t.Field(i).Name] = v.Field(i).Interface()
	}
	return data
}

// 特殊URL编码这个是POP特殊的一种规则，
// 即在一般的URLEncode后再增加三种字符替换：加号（+）替换成 %20、星号（*）替换成 %2A、%7E 替换回波浪号（~）
func SpecialUrlEncode(arg string) string {
	newString := url.QueryEscape(arg)
	newString = strings.Replace(newString, "+", "%20", -1)
	newString = strings.Replace(newString, "*", "%2A", -1)
	newString = strings.Replace(newString, "%7E", "~", -1)

	return newString
}

// 生成Length长度的数字
func RandNumber(length int) string {
	if length <= 0 {
		return ""
	}
	number := make([]string, 0)
	rand.Seed(time.Now().UnixNano())
	for i := 0; i < length; i++ {
		number = append(number, strconv.Itoa(rand.Intn(10)))
	}
	return strings.Join(number, "")
}
