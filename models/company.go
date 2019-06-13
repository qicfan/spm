package models

import (
	"github.com/jinzhu/gorm"
	"supermentor/db"
)

type Company struct {
	gorm.Model
	Name     string `gorm:"type:varchar(500)"` // 公司名称
	Logo     string `gorm:"type:varchar(500)"` // 公司logo
	Initial  string `gorm:"type:char(1)"`      // 大写首字母
	Category int    `gorm:"type:int"`          // 类别：1-寿险综合，2-财险，3-中介机构
}

// 表名
func (Company) TableName() string {
	return "company"
}

var cachedCompines = make([]Company, 0)

func GetAllCompanies() []Company {
	if len(cachedCompines) == 0 {
		db.Db.Find(&cachedCompines)
	}
	return cachedCompines
}

func GetCompanyById(companyId uint) *Company {
	company := &Company{}
	db.Db.Find(company, companyId)
	return company
}
