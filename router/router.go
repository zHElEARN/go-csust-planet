package router

import (
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "github.com/zHElEARN/go-csust-planet/docs"

	"github.com/zHElEARN/go-csust-planet/config"
	"github.com/zHElEARN/go-csust-planet/controller"
	"github.com/zHElEARN/go-csust-planet/middleware"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()

	v1 := r.Group("/v1")

	// 仅开发模式下启用
	if config.AppConfig.AppMode != "production" {
		utilGroup := v1.Group("/util")
		utilGroup.Use(middleware.AuthMiddleware())
		{
			utilGroup.GET("/hello", controller.Hello)
			utilGroup.GET("/electricity", controller.Electricity)
			utilGroup.GET("/profile", controller.Profile)
			utilGroup.POST("/push", controller.Push)
		}

		r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}

	taskGroup := v1.Group("/task")
	taskGroup.Use(middleware.AuthMiddleware())
	{
		taskGroup.POST("/electricity", controller.SyncElectricityTask)
	}

	authGroup := v1.Group("/auth")
	{
		authGroup.POST("/login", controller.Login)
	}

	configGroup := v1.Group("/config")
	{
		configGroup.GET("/announcements", controller.GetAnnouncements)
		configGroup.GET("/campus-map", controller.GetCampusMap)
		configGroup.GET("/app-versions", controller.GetAppVersions)
		configGroup.GET("/app-version/check", controller.CheckAppVersion)
	}

	r.NoRoute(controller.HandleNotFound)

	return r
}
