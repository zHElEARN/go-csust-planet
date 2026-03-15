package main

import (
	"log"

	"github.com/zHElEARN/go-csust-planet/config"
	"github.com/zHElEARN/go-csust-planet/router"
	"github.com/zHElEARN/go-csust-planet/utils/apns"
	"github.com/zHElEARN/go-csust-planet/worker"
)

// @title           go-csust-planet API
// @version         1.0
// @description     go-csust-planet 项目的 API 接口文档
// @host            localhost:8080
// @BasePath        /v1
func main() {
	config.InitConfig()
	config.InitDB()
	apns.InitAPNS()

	worker.StartElectricityPushWorker()

	r := router.SetupRouter()

	err := r.Run(":" + config.AppConfig.Port)
	if err != nil {
		log.Fatalf("服务器启动失败: %v", err)
	}
}
