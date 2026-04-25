package main

import (
	"log"

	"github.com/zHElEARN/go-csust-planet/config"
	"github.com/zHElEARN/go-csust-planet/controller"
	"github.com/zHElEARN/go-csust-planet/router"
	"github.com/zHElEARN/go-csust-planet/service"
	"github.com/zHElEARN/go-csust-planet/utils/apns"
	"github.com/zHElEARN/go-csust-planet/utils/campuscard"
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
	campuscard.InitBuildingStoreBlocking()

	authService := service.NewAuthService(
		config.DB,
		service.ProfileFetcherFunc(sso.GetUserProfile),
		service.TokenGeneratorFunc(jwt.GenerateToken),
	)
	electricityTaskService := service.NewElectricityTaskService(
		config.DB,
		service.BuildingResolverFunc(campuscard.GetBuildingByCampusName),
		nil,
	)
	adminAppVersionService := service.NewAdminAppVersionService(config.DB)
	pushConfig := service.DefaultElectricityPushConfig()
	electricityPushService := service.NewElectricityPushService(
		config.DB,
		service.BuildingResolverFunc(campuscard.GetBuildingByCampusName),
		service.ElectricityFetcherFunc(campuscard.GetElectricity),
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

	err := r.Run(":" + config.AppConfig.Port)
	if err != nil {
		log.Fatalf("[FATAL] 服务器启动失败: %v", err)
	}
}
