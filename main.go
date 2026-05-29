package main

import (
	"context"
	"log"
	"time"

	"github.com/zHElEARN/go-csust-planet/config"
	"github.com/zHElEARN/go-csust-planet/controller"
	"github.com/zHElEARN/go-csust-planet/router"
	"github.com/zHElEARN/go-csust-planet/service"
	"github.com/zHElEARN/go-csust-planet/utils/apns"
	"github.com/zHElEARN/go-csust-planet/utils/csustkit"
	"github.com/zHElEARN/go-csust-planet/utils/jwt"
	"github.com/zHElEARN/go-csust-planet/utils/sso"
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

	electricityInitCtx, electricityInitCancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer electricityInitCancel()
	electricityClient, err := csustkit.NewAuthenticatedElectricityClient(
		electricityInitCtx,
		config.AppConfig.CSUSTAuthserverUsername,
		config.AppConfig.CSUSTAuthserverPassword,
	)
	if err != nil {
		log.Fatalf("[FATAL] 校园卡系统登录失败: %v", err)
	}

	authService := service.NewAuthService(
		config.DB,
		service.ProfileFetcherFunc(sso.GetUserProfile),
		service.TokenGeneratorFunc(jwt.GenerateToken),
	)
	electricityTaskService := service.NewElectricityTaskService(
		config.DB,
		electricityClient,
		nil,
	)
	adminAppVersionService := service.NewAdminAppVersionService(config.DB)
	pushConfig := service.DefaultElectricityPushConfig()
	electricityPushService := service.NewElectricityPushService(
		config.DB,
		electricityClient,
		service.NotificationSenderFunc(apns.SendPushNotification),
		pushConfig,
	)

	handler := controller.NewHandler(controller.Dependencies{
		DB:                     config.DB,
		AuthService:            authService,
		ElectricityTaskService: electricityTaskService,
		AdminAppVersionService: adminAppVersionService,
	})

	worker.StartElectricityPushWorker(electricityPushService, service.DefaultWorkerTickInterval)

	r := router.SetupRouter(router.Dependencies{
		Handler:          handler,
		AppMode:          config.AppConfig.AppMode,
		SwaggerPassword:  config.AppConfig.SwaggerPassword,
		AdminBearerToken: config.AppConfig.AdminBearerToken,
	})

	err = r.Run(":" + config.AppConfig.Port)
	if err != nil {
		log.Fatalf("[FATAL] 服务器启动失败: %v", err)
	}
}
