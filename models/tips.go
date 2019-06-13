package models

import (
	"github.com/jinzhu/gorm"
	"time"
	"strconv"
	"supermentor/db"
	"math/rand"
)

type Tips struct {
	gorm.Model
	Content string `gorm:"type:varchar(1000)"`
	Remark string
	Category string
	StartTime int
	FinishTime int
}

func (Tips) TableName() string {
	return "tips"
}

// 取当前时间对应的一条tip
func GetTips() Tips {
	// 取当前时间 2006-01-02 15:04:05
	date,_ := strconv.Atoi(time.Now().Format("1504"))
	// 取这个区间的记录
	tips := make([]Tips, 0)
	if db.Db.Where("start_time<=? AND finish_time>=?", date, date).Find(&tips).RecordNotFound() {
		// 如果没有，则返回一个空tips
		return Tips{}
	}
	l := len(tips)
	if l == 0 {
		return Tips{}
	}
	if l == 1 {
		return tips[0]
	}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	i := r.Intn(l - 1)
	return tips[i]
}