package main

import (
	"log"

	"github.com/zHElEARN/go-csust-planet/config"
	"github.com/zHElEARN/go-csust-planet/controller"
	"github.com/zHElEARN/go-csust-planet/router"
	"github.com/zHElEARN/go-csust-planet/service"
)

// @title           go-csust-planet API
// @version         1.0
// @description     go-csust-planet 项目的 API 接口文档
// @host            localhost:8080
// @BasePath        /v1
func main() {
	config.InitConfig()
	config.InitDB()

	adminAppVersionService := service.NewAdminAppVersionService(config.DB)

	handler := controller.NewHandler(controller.Dependencies{
		DB:                     config.DB,
		AdminAppVersionService: adminAppVersionService,
	})

	r := router.SetupRouter(router.Dependencies{
		Handler:          handler,
		AppMode:          config.AppConfig.AppMode,
		SwaggerPassword:  config.AppConfig.SwaggerPassword,
		AdminBearerToken: config.AppConfig.AdminBearerToken,
	})

	if err := r.Run(":" + config.AppConfig.Port); err != nil {
		log.Fatalf("[FATAL] 服务器启动失败: %v", err)
	}
}
