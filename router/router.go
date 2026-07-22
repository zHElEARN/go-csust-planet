package router

import (
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "github.com/zHElEARN/go-csust-planet/docs"

	"github.com/zHElEARN/go-csust-planet/internal/announcement"
	"github.com/zHElEARN/go-csust-planet/internal/appversion"
	"github.com/zHElEARN/go-csust-planet/internal/campusmap"
	"github.com/zHElEARN/go-csust-planet/internal/health"
	"github.com/zHElEARN/go-csust-planet/internal/semestercalendar"
	"github.com/zHElEARN/go-csust-planet/middleware"
	"github.com/zHElEARN/go-csust-planet/utils/response"
)

type Dependencies struct {
	HealthHandler           *health.Handler
	AnnouncementHandler     *announcement.Handler
	AppVersionHandler       *appversion.Handler
	CampusMapHandler        *campusmap.Handler
	SemesterCalendarHandler *semestercalendar.Handler
	AppMode                 string
	SwaggerPassword         string
	AdminBearerToken        string
	CORSAllowedOrigins      []string
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
	r.Use(corsMiddleware(deps.CORSAllowedOrigins))

	r.GET("/healthz", deps.HealthHandler.Check)

	v1 := r.Group("/v1")

	swaggerGroup := r.Group("/swagger", gin.BasicAuth(gin.Accounts{
		"swagger": deps.SwaggerPassword,
	}))
	swaggerGroup.GET("/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	configGroup := v1.Group("/config")
	{
		configGroup.GET("/announcements", deps.AnnouncementHandler.GetAnnouncements)
		configGroup.GET("/campus-map", deps.CampusMapHandler.GetCampusMap)
		configGroup.GET("/app-versions", deps.AppVersionHandler.GetAppVersions)
		configGroup.GET("/app-versions/check", deps.AppVersionHandler.CheckAppVersion)
		configGroup.GET("/semester-calendars", deps.SemesterCalendarHandler.GetSemesterCalendars)
		configGroup.GET("/semester-calendars/:semester_code", deps.SemesterCalendarHandler.GetSemesterCalendarDetail)
	}

	adminGroup := v1.Group("/admin")
	adminGroup.Use(middleware.AdminAuthMiddleware(deps.AdminBearerToken))
	{
		adminGroup.GET("/announcements", deps.AnnouncementHandler.GetAdminAnnouncements)
		adminGroup.GET("/announcements/:id", deps.AnnouncementHandler.GetAdminAnnouncement)
		adminGroup.POST("/announcements", deps.AnnouncementHandler.CreateAnnouncement)
		adminGroup.PUT("/announcements/:id", deps.AnnouncementHandler.UpdateAnnouncement)
		adminGroup.DELETE("/announcements/:id", deps.AnnouncementHandler.DeleteAnnouncement)

		adminGroup.GET("/app-versions", deps.AppVersionHandler.GetAdminAppVersions)
		adminGroup.GET("/app-versions/:id", deps.AppVersionHandler.GetAdminAppVersion)
		adminGroup.POST("/app-versions", deps.AppVersionHandler.CreateAppVersion)
		adminGroup.PUT("/app-versions/:id", deps.AppVersionHandler.UpdateAppVersion)
		adminGroup.DELETE("/app-versions/:id", deps.AppVersionHandler.DeleteAppVersion)

		adminGroup.GET("/semester-calendars", deps.SemesterCalendarHandler.GetAdminSemesterCalendars)
		adminGroup.GET("/semester-calendars/:semester_code", deps.SemesterCalendarHandler.GetAdminSemesterCalendar)
		adminGroup.POST("/semester-calendars", deps.SemesterCalendarHandler.CreateSemesterCalendar)
		adminGroup.PUT("/semester-calendars/:semester_code", deps.SemesterCalendarHandler.UpdateSemesterCalendar)
		adminGroup.DELETE("/semester-calendars/:semester_code", deps.SemesterCalendarHandler.DeleteSemesterCalendar)
	}

	mountAdminFrontend(r, deps.AppMode)

	r.NoRoute(func(c *gin.Context) { response.ResponseError(c, 404, "找不到路由") })

	return r
}
