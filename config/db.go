package config

import (
	"fmt"
	"log"
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/zHElEARN/go-csust-planet/model"
)

var DB *gorm.DB

func InitDB() {
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=%s",
		AppConfig.DBHost, AppConfig.DBUser, AppConfig.DBPassword, AppConfig.DBName, AppConfig.DBPort, AppConfig.DBSSLMode, AppConfig.DBTimeZone,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.New(
			log.New(os.Stdout, "\r\n", log.LstdFlags),
			logger.Config{
				SlowThreshold:             200 * time.Millisecond,
				LogLevel:                  logger.Warn,
				IgnoreRecordNotFoundError: true,
				ParameterizedQueries:      true,
				Colorful:                  false,
			},
		),
	})
	if err != nil {
		log.Fatalf("[FATAL] 连接数据库失败: %v", err)
	}

	DB = db
	log.Println("[INFO] PostgreSQL 数据库连接成功")

	if err := AutoMigrate(db); err != nil {
		log.Fatalf("[FATAL] 数据库自动迁移失败: %v", err)
	}
	log.Println("[INFO] 数据库自动迁移完成")
}

func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&model.User{},
		&model.DeviceToken{},
		&model.ElectricityTask{},
		&model.Announcement{},
		&model.CampusMapFeature{},
		&model.AppVersion{},
		&model.SemesterCalendar{},
	)
}
