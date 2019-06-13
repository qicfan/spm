package db

import (
	"github.com/go-redis/redis"
	"supermentor/helpers"
)

var Redis *redis.Client

func ConnectRedis(host string, db int) *redis.Client {
	Redis = redis.NewClient(&redis.Options{
		Addr:     host,
		Password: "supermentor2018", // no password set
		DB:       db,                // use default DB
		PoolSize: 5,
	})
	// 设置日志
	redis.SetLogger(helpers.Log.Logger)
	_, err := Redis.Ping().Result()
	if err != nil {
		// 有错误
		helpers.Log.Error("redis连接失败: %+v", err)
		panic(err)
	}
	helpers.Log.Info("redis连接成功: %+v", Redis)
	return Redis
}
