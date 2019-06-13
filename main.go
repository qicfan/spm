package main

import (
	"flag"
	"fmt"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/kylelemons/go-gypsy/yaml"
	"github.com/robfig/cron"
	"net/http"
	"os"
	"strconv"
	"supermentor/controllers"
	"supermentor/db"
	"supermentor/helpers"
	"supermentor/models"
)

// 读取配置文件
// 连接数据库
// 连接redis
// 初始化web服务
func main() {
	configFile := flag.String("c", "/data/config.yml", "the config file")
	flag.Parse()
	// 读取配置文件
	config, _ := yaml.ReadFile(*configFile)
	initLogger(config)
	defer helpers.CloseLogger()
	initDb(config)
	defer db.Db.Close()
	// 初始化微信设置
	helpers.WechatAppId, _ = config.Get("wechat.appId")
	helpers.WechatAppSecret, _ = config.Get("wechat.appSecret")
	helpers.WechatMpAppId, _ = config.Get("wechatMp.appId")
	helpers.WechatMpAppSecret, _ = config.Get("wechatMp.appSecret")
	initRedis(config)
	defer db.Redis.Close()
	helpers.QiniuAK, _ = config.Get("qiniu.accessKey")
	helpers.QiniuSK, _ = config.Get("qiniu.secretKey")
	helpers.AliyunAK, _ = config.Get("aliyun.accessKey")
	helpers.AliyunSK, _ = config.Get("aliyun.secretKey")
	// 启动定时任务
	go initCron()
	// 初始化路由
	setRouter()
	// 启动web server
	host, _ := config.Get("web.host")
	err := http.ListenAndServe(host, nil) //设置监听的端口
	if err != nil {
		helpers.Log.Error("ListenAndServe: %v", err)
		os.Exit(1)
	}
}

// 设置路由
func setRouter() {
	routes := []controllers.SmRoute{
		controllers.SmRoute{
			Pattern:   "/wechat-auth",
			Handler:   controllers.LoginAction,
			NeedLogin: false,
			IsAdmin:   false,
		},
		controllers.SmRoute{
			Pattern:   "/check-android-update",
			Handler:   controllers.CheckAndroidUpdateAction,
			NeedLogin: false,
			IsAdmin:   false,
		},
		controllers.SmRoute{
			Pattern:   "/tip",
			Handler:   controllers.GetTip,
			NeedLogin: false,
			IsAdmin:   false,
		},
		controllers.SmRoute{
			Pattern:   "/user/info",
			Handler:   controllers.GetUserInfo,
			NeedLogin: true,
			IsAdmin:   false,
		},
		controllers.SmRoute{
			Pattern:   "/report/week",
			Handler:   controllers.GetWeekReport,
			NeedLogin: true,
			IsAdmin:   false,
		},
		controllers.SmRoute{
			Pattern:   "/report/data-columns",
			Handler:   controllers.GetReportDataColumn,
			NeedLogin: false,
			IsAdmin:   false,
		},
		controllers.SmRoute{
			Pattern:   "/report/update-data",
			Handler:   controllers.UpdateData,
			NeedLogin: true,
			IsAdmin:   false,
		},
		controllers.SmRoute{
			Pattern:   "/course/detail",
			Handler:   controllers.GetCourseDetail,
			NeedLogin: true,
			IsAdmin:   false,
		},
		controllers.SmRoute{
			Pattern:   "/course/study-users",
			Handler:   controllers.GetCourseStudyUsers,
			NeedLogin: true,
			IsAdmin:   false,
		},
		controllers.SmRoute{
			Pattern:   "/user/course",
			Handler:   controllers.GetUserRecommendCourse,
			NeedLogin: true,
			IsAdmin:   false,
		},
		controllers.SmRoute{
			Pattern:   "/course/progress",
			Handler:   controllers.UpdateCourseProgress,
			NeedLogin: true,
			IsAdmin:   false,
		},
		controllers.SmRoute{
			Pattern:   "/course/start",
			Handler:   controllers.StartCourse,
			NeedLogin: true,
			IsAdmin:   false,
		},
		controllers.SmRoute{
			Pattern:   "/abilities",
			Handler:   controllers.GetAbilities,
			NeedLogin: false,
			IsAdmin:   false,
		},
		controllers.SmRoute{
			Pattern:   "/user/report",
			Handler:   controllers.GetUserReport,
			NeedLogin: true,
			IsAdmin:   false,
		},
		controllers.SmRoute{
			Pattern:   "/report/courses",
			Handler:   controllers.GetReportUserCourse,
			NeedLogin: true,
			IsAdmin:   false,
		},
		controllers.SmRoute{
			Pattern:   "/user/action-count",
			Handler:   controllers.GetUserActionCount,
			NeedLogin: true,
			IsAdmin:   false,
		},
		controllers.SmRoute{
			Pattern:   "/course/questions",
			Handler:   controllers.GetCourseQuestion,
			NeedLogin: true,
			IsAdmin:   false,
		},
		controllers.SmRoute{
			Pattern:   "/course/save-questions",
			Handler:   controllers.SaveCourseQuestion,
			NeedLogin: true,
			IsAdmin:   true,
		},
		controllers.SmRoute{
			Pattern:   "/course/save-test",
			Handler:   controllers.SaveTestResult,
			NeedLogin: true,
			IsAdmin:   false,
		},
		controllers.SmRoute{
			Pattern:   "/course/tests",
			Handler:   controllers.GetCourseTests,
			NeedLogin: true,
			IsAdmin:   false,
		},
		controllers.SmRoute{
			Pattern:   "/user/update-name",
			Handler:   controllers.UpdateUserNickName,
			NeedLogin: true,
			IsAdmin:   false,
		},
		controllers.SmRoute{
			Pattern:   "/companies",
			Handler:   controllers.GetAllCompanies,
			NeedLogin: false,
			IsAdmin:   false,
		},
		controllers.SmRoute{
			Pattern:   "/user/update-company",
			Handler:   controllers.UpdateUserCompany,
			NeedLogin: true,
			IsAdmin:   false,
		},
		controllers.SmRoute{
			Pattern:   "/user/update-mobile",
			Handler:   controllers.UpdateUserMobile,
			NeedLogin: true,
			IsAdmin:   false,
		},
		controllers.SmRoute{
			Pattern:   "/user/tmp-auth",
			Handler:   controllers.TmpAuth,
			NeedLogin: false,
			IsAdmin:   false,
		},
		controllers.SmRoute{
			Pattern:   "/user/finish-guide",
			Handler:   controllers.UserFinishGuide,
			NeedLogin: true,
			IsAdmin:   false,
		},
		controllers.SmRoute{
			Pattern:   "/user/finish-evaluating",
			Handler:   controllers.UserFinishEvaluating,
			NeedLogin: true,
			IsAdmin:   false,
		},
		controllers.SmRoute{
			Pattern:   "/admin/device-id",
			Handler:   controllers.GetDeviceId,
			NeedLogin: false,
			IsAdmin:   true,
		},
		controllers.SmRoute{
			Pattern:   "/admin/login",
			Handler:   controllers.AdminLogin,
			NeedLogin: false,
			IsAdmin:   true,
		},
		controllers.SmRoute{
			Pattern:   "/admin/course-list",
			Handler:   controllers.AdminCourseList,
			NeedLogin: true,
			IsAdmin:   true,
		},
		controllers.SmRoute{
			Pattern:   "/uptoken",
			Handler:   controllers.GetUpToken,
			NeedLogin: false,
			IsAdmin:   true,
		},
		controllers.SmRoute{
			Pattern:   "/admin/course-add",
			Handler:   controllers.AdminAddCourse,
			NeedLogin: true,
			IsAdmin:   true,
		},
		controllers.SmRoute{
			Pattern:   "/send-code",
			Handler:   controllers.SendCode,
			NeedLogin: true,
			IsAdmin:   false,
		},
		controllers.SmRoute{
			Pattern:   "/team/search-mobile",
			Handler:   controllers.SearchTeamByMobile,
			NeedLogin: true,
			IsAdmin:   false,
		},
		controllers.SmRoute{
			Pattern:   "/team/leave",
			Handler:   controllers.LeaveTeam,
			NeedLogin: true,
			IsAdmin:   false,
		},
		controllers.SmRoute{
			Pattern:   "/team/delete-user",
			Handler:   controllers.DeleteUserFromTeam,
			NeedLogin: true,
			IsAdmin:   false,
		},
		controllers.SmRoute{
			Pattern:   "/team/apply",
			Handler:   controllers.ApplyTeam,
			NeedLogin: true,
			IsAdmin:   false,
		},
		controllers.SmRoute{
			Pattern:   "/team/apply-pass",
			Handler:   controllers.TeamApplyPass,
			NeedLogin: true,
			IsAdmin:   false,
		},
		controllers.SmRoute{
			Pattern:   "/team/apply-reject",
			Handler:   controllers.TeamApplyReject,
			NeedLogin: true,
			IsAdmin:   false,
		},
		controllers.SmRoute{
			Pattern:   "/team/members",
			Handler:   controllers.TeamMembers,
			NeedLogin: true,
			IsAdmin:   false,
		},
		controllers.SmRoute{
			Pattern:   "/team/info",
			Handler:   controllers.GetTeamInfo,
			NeedLogin: true,
			IsAdmin:   false,
		},
		controllers.SmRoute{
			Pattern:   "/team/applies",
			Handler:   controllers.TeamApplies,
			NeedLogin: true,
			IsAdmin:   false,
		},
		controllers.SmRoute{
			Pattern:   "/team/create",
			Handler:   controllers.CreateTeam,
			NeedLogin: true,
			IsAdmin:   false,
		},
		controllers.SmRoute{
			Pattern:   "/config",
			Handler:   controllers.GetAppConfig,
			NeedLogin: false,
			IsAdmin:   false,
		},
		controllers.SmRoute{
			Pattern:   "/pics",
			Handler:   controllers.GetPics,
			NeedLogin: false,
			IsAdmin:   false,
		},
		controllers.SmRoute{
			Pattern:   "/common/enc",
			Handler:   controllers.Encrypt,
			NeedLogin: true,
			IsAdmin:   false,
		},
		controllers.SmRoute{
			Pattern:   "/wechat/auth",
			Handler:   controllers.WechatAuth,
			NeedLogin: false,
			IsAdmin:   false,
		},
		controllers.SmRoute{
			Pattern:   "/wechat/callback",
			Handler:   controllers.WechatCallback,
			NeedLogin: false,
			IsAdmin:   false,
		},
	}
	for _, route := range routes {
		http.Handle(route.Pattern, route)
	}
	helpers.Log.Info("init http success")
}

func initDb(config *yaml.File) {
	username, _ := config.Get("db.username")
	password, _ := config.Get("db.password")
	dbHost, _ := config.Get("db.host")
	database, _ := config.Get("db.database")
	// 初始化数据库配置文件
	db.SetConfig(username, password, dbHost, database)
	// 建立数据库连接
	db.GetDb()
	// 设置数据库日志
	db.Db.SetLogger(helpers.Log.Logger)
	db.Db.LogMode(true)
	fmt.Println("init db success")
}

func initLogger(config *yaml.File) {
	var logFile string
	var logErr error
	// 初始化日志
	if logFile, logErr = config.Get("log.file"); logErr != nil {
		logFile = "./supermentor.log"
	}
	helpers.Log = helpers.CreateLogger(logFile)
	fmt.Println("init logger success")
}

func initRedis(config *yaml.File) {
	host, _ := config.Get("redis.host")
	redisDb, _ := config.Get("redis.db")
	redisDbInt, _ := strconv.Atoi(redisDb)
	db.ConnectRedis(host, redisDbInt)
	fmt.Println("init redis success")
}

// 定义定时任务
func initCron() {
	c := cron.New()
	// 每日21点计算拜访量排行榜
	c.AddFunc("0 0 21 * * *", models.CloseDayReportCallRank)
	// 每周日晚23:59:59点，结算周报
	c.AddFunc("59 59 23 * * 0", models.CloseWeekReport)
	helpers.Log.Info("init cron success")
	c.Start()
}
