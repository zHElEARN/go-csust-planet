package main

import (
	"context"
	"log"

	"github.com/zHElEARN/go-csust-planet/config"
	"github.com/zHElEARN/go-csust-planet/internal/announcement"
	"github.com/zHElEARN/go-csust-planet/internal/appversion"
	"github.com/zHElEARN/go-csust-planet/internal/campusmap"
	"github.com/zHElEARN/go-csust-planet/internal/health"
	"github.com/zHElEARN/go-csust-planet/internal/postgres"
	"github.com/zHElEARN/go-csust-planet/internal/semestercalendar"
	"github.com/zHElEARN/go-csust-planet/router"
)

// @title           go-csust-planet API
// @version         1.0
// @description     go-csust-planet 项目的 API 接口文档
// @host            localhost:8080
// @BasePath        /
func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("[FATAL] 加载配置失败: %v", err)
	}
	db, err := postgres.Open(cfg)
	if err != nil {
		log.Fatalf("[FATAL] 连接数据库失败: %v", err)
	}
	migrationVersion, err := postgres.Migrate(context.Background(), db)
	if err != nil {
		log.Fatalf("[FATAL] 数据库版本迁移失败: %v", err)
	}
	log.Printf("[INFO] PostgreSQL 数据库连接成功，当前迁移版本: %d", migrationVersion)

	announcementHandler := announcement.NewHandler(announcement.NewService(announcement.NewPostgresRepository(db)))
	legacyAppVersionHandler := appversion.NewHandler(appversion.NewService(appversion.NewLegacyPostgresRepository(db)))
	appVersionHandler := appversion.NewHandler(appversion.NewService(appversion.NewPostgresRepository(db)))
	campusMapHandler := campusmap.NewHandler(campusmap.NewService(campusmap.NewPostgresRepository(db)))
	semesterCalendarHandler := semestercalendar.NewHandler(semestercalendar.NewService(semestercalendar.NewPostgresRepository(db)))

	r := router.SetupRouter(router.Dependencies{
		HealthHandler:           health.NewHandler(db),
		AnnouncementHandler:     announcementHandler,
		LegacyAppVersionHandler: legacyAppVersionHandler,
		AppVersionHandler:       appVersionHandler,
		CampusMapHandler:        campusMapHandler,
		SemesterCalendarHandler: semesterCalendarHandler,
		AppMode:                 cfg.AppMode,
		SwaggerPassword:         cfg.SwaggerPassword,
		AdminBearerToken:        cfg.AdminBearerToken,
		CORSAllowedOrigins:      cfg.CORSAllowedOrigins,
		BusinessRequestTimeout:  cfg.BusinessRequestTimeout,
	})

	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatalf("[FATAL] 服务器启动失败: %v", err)
	}
}
