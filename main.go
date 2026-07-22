package main

import (
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
// @BasePath        /v1
func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("[FATAL] 加载配置失败: %v", err)
	}
	db, err := postgres.Open(cfg)
	if err != nil {
		log.Fatalf("[FATAL] 连接数据库失败: %v", err)
	}
	if err := postgres.AutoMigrate(db, &announcement.Entity{}, &appversion.Entity{}, &campusmap.Entity{}, &semestercalendar.Entity{}); err != nil {
		log.Fatalf("[FATAL] 数据库自动迁移失败: %v", err)
	}
	log.Println("[INFO] PostgreSQL 数据库连接成功，自动迁移完成")

	announcementHandler := announcement.NewHandler(announcement.NewService(announcement.NewPostgresRepository(db)))
	appVersionHandler := appversion.NewHandler(appversion.NewService(appversion.NewPostgresRepository(db)))
	campusMapHandler := campusmap.NewHandler(campusmap.NewService(campusmap.NewPostgresRepository(db)))
	semesterCalendarHandler := semestercalendar.NewHandler(semestercalendar.NewService(semestercalendar.NewPostgresRepository(db)))

	r := router.SetupRouter(router.Dependencies{
		HealthHandler:           health.NewHandler(db),
		AnnouncementHandler:     announcementHandler,
		AppVersionHandler:       appVersionHandler,
		CampusMapHandler:        campusMapHandler,
		SemesterCalendarHandler: semesterCalendarHandler,
		AppMode:                 cfg.AppMode,
		SwaggerPassword:         cfg.SwaggerPassword,
		AdminBearerToken:        cfg.AdminBearerToken,
		CORSAllowedOrigins:      cfg.CORSAllowedOrigins,
	})

	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatalf("[FATAL] 服务器启动失败: %v", err)
	}
}
