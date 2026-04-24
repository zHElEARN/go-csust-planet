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
	if config.AppConfig.AppMode == "production" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	r := gin.Default()

	v1 := r.Group("/v1")

	swaggerGroup := r.Group("/swagger", gin.BasicAuth(gin.Accounts{
		"swagger": config.AppConfig.SwaggerPassword,
	}))
	swaggerGroup.GET("/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

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
		configGroup.GET("/app-versions/check", controller.CheckAppVersion)
		configGroup.GET("/semester-calendars", controller.GetSemesterCalendars)
		configGroup.GET("/semester-calendars/:semester_code", controller.GetSemesterCalendarDetail)
	}

	r.NoRoute(controller.HandleNotFound)

	return r
}
