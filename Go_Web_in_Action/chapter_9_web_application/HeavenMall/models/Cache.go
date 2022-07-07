package models

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/astaxie/beego/client/cache"
	_ "github.com/astaxie/beego/client/cache/redis"
	"github.com/astaxie/beego/core/logs"
	"github.com/astaxie/beego/server/web"
)

var redisClient cache.Cache
var enableRedis, _ = web.AppConfig.Bool("enableRedis")
var redisTime, _ = web.AppConfig.Int("redisTime")
var YzmClient cache.Cache
var logger = logs.GetBeeLogger()

func init() {
	if enableRedis {
		redisKey, _ := web.AppConfig.String("redisKey")
		redisConn, _ := web.AppConfig.String("redisConn")
		redisDbNum, _ := web.AppConfig.String("redisDbNum")
		redisPwd, _ := web.AppConfig.String("redisPwd")
		config := map[string]string{
			"key":      redisKey,
			"conn":     redisConn,
			"dbNum":    redisDbNum,
			"password": redisPwd,
		}
		bytes, _ := json.Marshal(config)

		redisClient, err = cache.NewCache("redis", string(bytes))
		YzmClient, _ = cache.NewCache("redis", string(bytes))
		if err != nil {
			logger.Error("failed to connect Redis")
		} else {
			logger.Info("connect to Redis successfully!")
		}
	}
}

type cacheDb struct{}

var CacheDb = &cacheDb{}

func (c cacheDb) Set(ctx context.Context, key string, value interface{}) {
	if enableRedis {
		bytes, _ := json.Marshal(value)
		redisClient.Put(ctx, key, string(bytes), time.Second*time.Duration(redisTime))
	}
}

func (c cacheDb) Get(ctx context.Context, key string, obj interface{}) bool {
	if enableRedis {
		if redisStr, err := redisClient.Get(ctx, key); err != nil {
			fmt.Println("read data from redis ...")
			redisValue, ok := redisStr.([]uint8)
			if !ok {
				fmt.Println("failed to get data from redis")
				return false
			}
			json.Unmarshal([]byte(redisValue), obj)
			return true
		}
		return false
	}
	return false
}
