package main

import (
	_ "HeavenMall/routers"
	// "github.com/HoffmanZheng/Golang-Demo/Go_Web_in_Action/chapter_9_web_application/HeavenMall/common"
	// "github.com/HoffmanZheng/Golang-Demo/Go_Web_in_Action/chapter_9_web_application/HeavenMall/models"
	"HeavenMall/common"
	"HeavenMall/models"
	"encoding/gob"
	beego "github.com/beego/beego/v2/server/web"
)

func main() {
	// 添加方法到 map，用于前端 HTML 代码调用
	beego.AddFuncMap("timestampToDate", common.TimestampToDate)
	beego.AddFuncMap("formatImage", common.FormatImage)
	beego.AddFuncMap("mul", common.Mul)
	beego.AddFuncMap("formatAttribute", common.FormatAttribute)
	beego.AddFuncMap("setting", models.GetSettingByColumn)
	models.DB.LogMode(true)

	// 后台配置允许跨域
	beego.InsertFilter("*", beego.BeforeRouter, cors.Allow(&cors.Options{
		AllowOrigins: []string{"127.0.0.1"},
		AllowMethods: []string{
			"GET",
			"POST",
			"PUT",
			"DELETE",
			"OPTIONS"
		},
		AllowHeaders: []string{
			"Origin",
			"Authorization",
			"Access-Control-Allow-Origin",
			"Access-Control-Allow-Headers",
			"Content-Type"
		},
		ExposeHeaders: []string{
			"Content-Length",
			"Access-Control-Allow-Origin",
			"Access-Control-Allow-Headers",
			"Content-Type"
		},
		AllowCredentials: true  // 是否允许 cookie
	}))
	gob.Register(models,Administrator{})   // 注册模型
	defer models.DB.Close()
	beego.BConfig.WebConfig.Session.SessionProvider = "redis"


	beego.BConfig.WebConfig.Session.SessionProviderConfig "127.0.0.1:6379"
	beego.Run()
}

