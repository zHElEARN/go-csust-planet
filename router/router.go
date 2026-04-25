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

	r := gin.New()
	r.Use(gin.LoggerWithConfig(gin.LoggerConfig{
		SkipPaths: []string{"/healthz"},
	}))
	r.Use(gin.Recovery())

	r.GET("/healthz", controller.HealthCheck)

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

	adminGroup := v1.Group("/admin")
	adminGroup.Use(middleware.AdminAuthMiddleware())
	{
		adminGroup.GET("/announcements", controller.GetAdminAnnouncements)
		adminGroup.GET("/announcements/:id", controller.GetAdminAnnouncement)
		adminGroup.POST("/announcements", controller.CreateAnnouncement)
		adminGroup.PUT("/announcements/:id", controller.UpdateAnnouncement)
		adminGroup.DELETE("/announcements/:id", controller.DeleteAnnouncement)

		adminGroup.GET("/app-versions", controller.GetAdminAppVersions)
		adminGroup.GET("/app-versions/:id", controller.GetAdminAppVersion)
		adminGroup.POST("/app-versions", controller.CreateAppVersion)
		adminGroup.PUT("/app-versions/:id", controller.UpdateAppVersion)
		adminGroup.DELETE("/app-versions/:id", controller.DeleteAppVersion)
	}

	r.NoRoute(controller.HandleNotFound)

	return r
}
