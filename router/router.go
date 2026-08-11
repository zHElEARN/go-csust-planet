package router

import (
	"encoding/json"
	"time"

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

type accessLogEntry struct {
	Time     string `json:"time"`
	Status   int    `json:"status"`
	Latency  string `json:"latency"`
	ClientIP string `json:"client_ip"`
	Method   string `json:"method"`
	Path     string `json:"path"`
	UA       string `json:"ua"`
	BodySize int    `json:"body_size"`
}

func jsonLogFormatter(param gin.LogFormatterParams) string {
	entry := accessLogEntry{
		Time:     param.TimeStamp.Format(time.RFC3339),
		Status:   param.StatusCode,
		Latency:  param.Latency.String(),
		ClientIP: param.ClientIP,
		Method:   param.Method,
		Path:     param.Path,
		UA:       param.Request.UserAgent(),
		BodySize: param.BodySize,
	}
	b, err := json.Marshal(entry)
	if err != nil {
		return ""
	}
	return string(b) + "\n"
}

type Dependencies struct {
	HealthHandler           *health.Handler
	AnnouncementHandler     *announcement.Handler
	LegacyAppVersionHandler *appversion.Handler
	AppVersionHandler       *appversion.Handler
	CampusMapHandler        *campusmap.Handler
	SemesterCalendarHandler *semestercalendar.Handler
	AppMode                 string
	SwaggerPassword         string
	AdminBearerToken        string
	CORSAllowedOrigins      []string
	BusinessRequestTimeout  time.Duration
}

func SetupRouter(deps Dependencies) *gin.Engine {
	if deps.AppMode == "production" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	r := gin.New()
	r.Use(gin.LoggerWithConfig(gin.LoggerConfig{
		Formatter: jsonLogFormatter,
		SkipPaths: []string{"/healthz"},
	}))
	r.Use(gin.Recovery())
	r.Use(corsMiddleware(deps.CORSAllowedOrigins))

	r.GET("/healthz", deps.HealthHandler.Check)

	v1 := r.Group("/v1")
	v1.Use(requestTimeout(deps.BusinessRequestTimeout))

	swaggerGroup := r.Group("/swagger", gin.BasicAuth(gin.Accounts{
		"swagger": deps.SwaggerPassword,
	}))
	swaggerGroup.GET("/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	configGroup := v1.Group("/config")
	{
		configGroup.GET("/announcements", deps.AnnouncementHandler.GetAnnouncements)
		configGroup.GET("/campus-map", deps.CampusMapHandler.GetCampusMap)
		configGroup.GET("/app-versions", deps.LegacyAppVersionHandler.GetAppVersions)
		configGroup.GET("/app-versions/check", deps.LegacyAppVersionHandler.CheckAppVersion)
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

		adminGroup.GET("/app-versions", deps.LegacyAppVersionHandler.GetAdminAppVersions)
		adminGroup.GET("/app-versions/:id", deps.LegacyAppVersionHandler.GetAdminAppVersion)
		adminGroup.POST("/app-versions", deps.LegacyAppVersionHandler.CreateAppVersion)
		adminGroup.PUT("/app-versions/:id", deps.LegacyAppVersionHandler.UpdateAppVersion)
		adminGroup.DELETE("/app-versions/:id", deps.LegacyAppVersionHandler.DeleteAppVersion)

		adminGroup.GET("/semester-calendars", deps.SemesterCalendarHandler.GetAdminSemesterCalendars)
		adminGroup.GET("/semester-calendars/:semester_code", deps.SemesterCalendarHandler.GetAdminSemesterCalendar)
		adminGroup.POST("/semester-calendars", deps.SemesterCalendarHandler.CreateSemesterCalendar)
		adminGroup.PUT("/semester-calendars/:semester_code", deps.SemesterCalendarHandler.UpdateSemesterCalendar)
		adminGroup.DELETE("/semester-calendars/:semester_code", deps.SemesterCalendarHandler.DeleteSemesterCalendar)
	}

	v2 := r.Group("/v2")
	v2.Use(requestTimeout(deps.BusinessRequestTimeout))

	v2ConfigGroup := v2.Group("/config")
	{
		v2ConfigGroup.GET("/app-versions", deps.AppVersionHandler.GetAppVersions)
		v2ConfigGroup.GET("/app-versions/check", deps.AppVersionHandler.CheckAppVersion)
	}

	v2AdminGroup := v2.Group("/admin")
	v2AdminGroup.Use(middleware.AdminAuthMiddleware(deps.AdminBearerToken))
	{
		v2AdminGroup.GET("/app-versions", deps.AppVersionHandler.GetAdminAppVersions)
		v2AdminGroup.GET("/app-versions/:id", deps.AppVersionHandler.GetAdminAppVersion)
		v2AdminGroup.POST("/app-versions", deps.AppVersionHandler.CreateAppVersion)
		v2AdminGroup.PUT("/app-versions/:id", deps.AppVersionHandler.UpdateAppVersion)
		v2AdminGroup.DELETE("/app-versions/:id", deps.AppVersionHandler.DeleteAppVersion)
	}

	mountAdminFrontend(r, deps.AppMode)

	r.NoRoute(func(c *gin.Context) { response.ResponseError(c, 404, "找不到路由") })

	return r
}
