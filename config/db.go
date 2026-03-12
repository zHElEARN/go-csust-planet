package config

import (
	"fmt"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/zHElEARN/go-csust-planet/model"
)

var DB *gorm.DB

func InitDB() {
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=%s",
		AppConfig.DBHost, AppConfig.DBUser, AppConfig.DBPassword, AppConfig.DBName, AppConfig.DBPort, AppConfig.DBSSLMode, AppConfig.DBTimeZone,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("连接数据库失败: %v", err)
	}

	DB = db
	log.Println("PostgreSQL 数据库连接成功")

	autoMigrate(db)
}

func autoMigrate(db *gorm.DB) {
	err := db.AutoMigrate(
		&model.User{},
		&model.DeviceToken{},
		&model.ElectricityTask{},
		&model.Announcement{},
		&model.CampusMapFeature{},
		&model.AppVersion{},
	)
	if err != nil {
		log.Fatalf("数据库自动迁移失败: %v", err)
	}
	log.Println("数据库自动迁移完成")
}
