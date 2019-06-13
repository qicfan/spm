package db

import (
	"github.com/jinzhu/gorm"
	"fmt"
)

var Db *gorm.DB
var dbConfig map[string]string

// 获取一个数据库连接
func GetDb() *gorm.DB {
	dsn := fmt.Sprintf("%s:%s@%s/%s?charset=utf8&parseTime=True&loc=Local", dbConfig["username"], dbConfig["password"], dbConfig["host"], dbConfig["database"])
	Db, _ = gorm.Open("mysql", dsn)
	// 设置连接池的属性
	Db.DB().SetMaxOpenConns(100)      // 最大连接数
	Db.DB().SetMaxOpenConns(5)        // 启动时开启的连接数
	Db.DB().SetConnMaxLifetime(28700) // 连接空闲时长，小于mysql.wait_timeout
	Db.DB().SetMaxIdleConns(5)        // 最小连接数
	return Db
}

// 设置数据库连接参数
func SetConfig(username string, password string, host string, database string) {
	dbConfig = make(map[string]string)
	dbConfig["username"] = username
	dbConfig["password"] = password
	dbConfig["host"] = host
	dbConfig["database"] = database
}
