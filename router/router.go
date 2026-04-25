package router

import (
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "github.com/zHElEARN/go-csust-planet/docs"

	"github.com/zHElEARN/go-csust-planet/controller"
	"github.com/zHElEARN/go-csust-planet/middleware"
)

type Dependencies struct {
	Handler          *controller.Handler
	AppMode          string
	SwaggerPassword  string
	AdminBearerToken string
}

func SetupRouter(deps Dependencies) *gin.Engine {
	if deps.AppMode == "production" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	r := gin.New()
	r.Use(gin.LoggerWithConfig(gin.LoggerConfig{
		SkipPaths: []string{"/healthz"},
	}))
	r.Use(gin.Recovery())

	r.GET("/healthz", deps.Handler.HealthCheck)

	v1 := r.Group("/v1")

	swaggerGroup := r.Group("/swagger", gin.BasicAuth(gin.Accounts{
		"swagger": deps.SwaggerPassword,
	}))
	swaggerGroup.GET("/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	taskGroup := v1.Group("/task")
	taskGroup.Use(middleware.AuthMiddleware())
	{
		taskGroup.POST("/electricity", deps.Handler.SyncElectricityTask)
	}

	authGroup := v1.Group("/auth")
	{
		authGroup.POST("/login", deps.Handler.Login)
	}

	configGroup := v1.Group("/config")
	{
		configGroup.GET("/announcements", deps.Handler.GetAnnouncements)
		configGroup.GET("/campus-map", deps.Handler.GetCampusMap)
		configGroup.GET("/app-versions", deps.Handler.GetAppVersions)
		configGroup.GET("/app-versions/check", deps.Handler.CheckAppVersion)
		configGroup.GET("/semester-calendars", deps.Handler.GetSemesterCalendars)
		configGroup.GET("/semester-calendars/:semester_code", deps.Handler.GetSemesterCalendarDetail)
	}

	adminGroup := v1.Group("/admin")
	adminGroup.Use(middleware.AdminAuthMiddleware(deps.AdminBearerToken))
	{
		adminGroup.GET("/announcements", deps.Handler.GetAdminAnnouncements)
		adminGroup.GET("/announcements/:id", deps.Handler.GetAdminAnnouncement)
		adminGroup.POST("/announcements", deps.Handler.CreateAnnouncement)
		adminGroup.PUT("/announcements/:id", deps.Handler.UpdateAnnouncement)
		adminGroup.DELETE("/announcements/:id", deps.Handler.DeleteAnnouncement)

		adminGroup.GET("/app-versions", deps.Handler.GetAdminAppVersions)
		adminGroup.GET("/app-versions/:id", deps.Handler.GetAdminAppVersion)
		adminGroup.POST("/app-versions", deps.Handler.CreateAppVersion)
		adminGroup.PUT("/app-versions/:id", deps.Handler.UpdateAppVersion)
		adminGroup.DELETE("/app-versions/:id", deps.Handler.DeleteAppVersion)
	}

	r.NoRoute(controller.HandleNotFound)

	return r
}
