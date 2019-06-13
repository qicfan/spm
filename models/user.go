package models

import (
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/silenceper/wechat"
	"strconv"
	"supermentor/db"
	"supermentor/helpers"
)

type User struct {
	gorm.Model
	Nickname          string            `gorm:"type:varchar(64)"` // 昵称
	Mobile            string            `gorm:"type:varchar(11)"` // 手机号
	Password          string            `gorm:"type:varchar(128)" json:"-"`
	Avatar            string            `gorm:"type:varchar(100)"`   // 头像
	Sex               uint              `gorm:"type:int" json:"sex"` // 性别
	LastLoginAt       uint              `gorm:"type:int"`            // 最后一次登录时间
	CompanyId         uint              // 公司Id
	IsNew             int               // 是否完成新手引导，0-未完成，1-完成新手引导，2-完成新手引导+评测
	UserThirdPlatform UserThirdPlatform // 用户绑定的第三方平台账号
	UserCompany       *Company          `gorm:"ForeignKey:CompanyId"` // 关联的公司
	BelongTeam        *Team             // 我属于的团队
	MyTeam            *Team             // 我创建的团队

	WeekReport *Report // 关联的本周周报
	EncryptID  string  `gorm:"-"` // 加密的id，用于外部传播
}

// 表名
func (User) TableName() string {
	return "user"
}

// 注册微信用户
func (user *User) CreateWechatUser(wechatUser *helpers.WechatUserStruct) error {
	tx := db.Db.Begin()
	if wechatUser.UniId != "" {
		// 根据uniid查询是否存在用户
		uniPlatformUser := UserThirdPlatform{}
		if !db.Db.Where("uniid=?", wechatUser.UniId).Find(&uniPlatformUser).RecordNotFound() {
			// 如果存在
			user.Model.ID = uniPlatformUser.UserID
		}
	}
	user.Nickname = wechatUser.NickName
	user.Avatar = wechatUser.Avatar
	user.Sex = wechatUser.Sex
	user.IsNew = 1

	if user.Model.ID == 0 {
		// 新建用户
		createErr := tx.Create(&user).Error
		if createErr != nil {
			helpers.Log.Error("create-user: %v", createErr)
			// 创建失败
			tx.Rollback()
			return createErr
		}
	}
	userThird := UserThirdPlatform{
		Platform:       "wechat",
		PlatformUserId: wechatUser.OpenId,
		UserID:         user.Model.ID,
		Avatar:         wechatUser.Avatar,
		City:           wechatUser.City,
		Province:       wechatUser.Province,
		Country:        wechatUser.Country,
		Uniid:          wechatUser.UniId,
		Nickname:       wechatUser.NickName,
	}
	tx.NewRecord(userThird)
	tErr := tx.Create(&userThird).Error
	if tErr != nil {
		helpers.Log.Error("create-user-third: %v", tErr)
		tx.Rollback()
		return tErr
	}
	tx.Commit()
	return nil
}

// 更新昵称
func (user *User) UpdateName(newName string) bool {
	if user.Nickname == newName {
		return false
	}
	db.Db.Model(user).Update("nickname", newName)
	user.Nickname = newName
	RecordUserAction(user.ID, UserUpdateInfoAction, "修改昵称")
	return true
}

// 更新公司
func (user *User) UpdateCompany(company *Company) {
	db.Db.Model(&User{}).Where("id=?", user.ID).Update("company_id", company.ID)
	user.CompanyId = company.ID
	user.UserCompany = company
	RecordUserAction(user.ID, UserUpdateInfoAction, "修改所属公司")
}

// 更新手机号
func (user *User) UpdateMobile(mobile string) {
	if user.Mobile != "" {
		// 如果已经有了，则不能更新，后期开发换绑功能
		return
	}
	db.Db.Model(&User{}).Where("id=?", user.ID).Update("mobile", mobile)
	user.Mobile = mobile
	RecordUserAction(user.ID, UserUpdateInfoAction, "绑定手机号")
}

// 使用设备id创建用户
func (user *User) CreateDeviceUser(did string) error {
	tx := db.Db.Begin()
	user.Nickname = "临时用户"
	user.Avatar = "assets/imgs/default-avatar.jpeg"
	user.Sex = 2
	tx.NewRecord(user)
	createErr := tx.Create(&user).Error
	if createErr != nil {
		helpers.Log.Error("create-user: %v", createErr)
		// 创建失败
		tx.Rollback()
		return createErr
	}
	userThird := UserThirdPlatform{
		Platform:       "device",
		PlatformUserId: did,
		UserID:         user.Model.ID,
		Avatar:         "",
		City:           "",
		Province:       "",
		Country:        "",
	}
	tx.NewRecord(userThird)
	tErr := tx.Create(&userThird).Error
	if tErr != nil {
		helpers.Log.Error("create-user-third: %v", tErr)
		tx.Rollback()
		return tErr
	}
	tx.Commit()
	return nil
}

// 完成新手引导
func (user *User) FinishGuide() {
	db.Db.Model(user).Update("is_new", 1)
	user.IsNew = 1
	RecordUserAction(user.ID, UserUpdateInfoAction, "完成新手引导")
}

// 完成评测
func (user *User) FinishEvaluating() {
	db.Db.Model(user).Update("is_new", 2)
	user.IsNew = 2
	RecordUserAction(user.ID, UserUpdateInfoAction, "完成评测")
}

// 查询用户所属的团队
func (user *User) GetBelongTeam() *User {
	tu := &TeamUser{}
	user.BelongTeam = &Team{}
	if db.Db.Where("user_id=?", user.ID).Find(tu).RecordNotFound() {
		return user
	}
	// 查询team
	db.Db.Find(user.BelongTeam, tu.TeamID)
	return user
}

// 查询用户直辖的团队
func (user *User) GetUserTeam() *User {
	user.MyTeam = GetOwnerTeam(user.ID)
	return user
}

// 根据用户id查询用户信息
func GetUserById(userId uint, hasTeam bool) (*User, error) {
	user := &User{}
	if db.Db.Preload("UserCompany").First(user, userId).RecordNotFound() || db.Db.Error != nil {
		helpers.Log.Error("user: %v", db.Db.Error)
		// 如果没有，则返回Nil
		return user, db.Db.Error
	}
	if hasTeam == true {
		// 查询所属团队和直辖团队
		user.GetBelongTeam().GetUserTeam()
	}
	uId := strconv.Itoa(int(user.ID))
	encryptIDByte, err := helpers.DesEncrypt([]byte(uId), []byte("sm201807"))
	fmt.Printf("%v", err)
	user.EncryptID = string(encryptIDByte[:])
	return user, nil
}

// 根据openid查询用户
func GetUserByPlatformId(platform string, platformUserId string) (*User, error) {
	var user = &User{}
	var userThird UserThirdPlatform
	if db.Db.First(&userThird, "platform = ? AND platform_user_id = ?", platform, platformUserId).RecordNotFound() {
		// 如果不存在或者有错误
		helpers.Log.Error("user: %v", db.Db.Error)
		return user, nil
	}
	return GetUserById(userThird.UserID, false)
}

// 通过手机号查询用户
func GetUserByMobile(mobile string, hasTeam bool) *User {
	user := &User{}
	db.Db.Preload("UserCompany").Where("mobile=?", mobile).First(user)
	if user.ID > 0 && hasTeam == true {
		// 查询所属团队和直辖团队
		user.GetBelongTeam().GetUserTeam()
	}
	return user
}

// 离开所属团队
func (user *User) LeaveTeam() bool {
	// 删除团队关系
	tu := &TeamUser{}
	if db.Db.Where("team_id=? AND user_id=?", user.BelongTeam.ID, user.ID).Find(tu).RecordNotFound() {
		return false
	}
	// 删除
	if err := db.Db.Delete(tu).Error; err != nil {
		helpers.Log.Warning("用户【%d】离开团队失败: %v", user.ID, err)
		return false
	}
	// 更新团队的成员人数
	user.BelongTeam.DecreaseMemberCount(1)
	if user.MyTeam.ID > 0 {
		// 如果有自己的团队，则更新parent_id = 0
		user.MyTeam.LeaveParent()
	}
	return true
}

// 查找团队申请列表
func (user *User) TeamApplies(page int) []TeamApply {
	pageSize := 100
	offset := (page - 1) * pageSize
	tas := make([]TeamApply, 0)
	where := fmt.Sprintf("(user_id=%d AND status>1)", user.ID)
	if user.MyTeam.ID > 0 {
		where = fmt.Sprintf("%s OR team_id=%d", where, user.MyTeam.ID)
	}
	db.Db.Where(where).Order("id DESC").Limit(pageSize).Offset(offset).Preload("ApplyUser").Preload("ApplyTeam").Find(&tas)
	return tas
}
