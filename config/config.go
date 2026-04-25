package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DBHost             string
	DBPort             string
	DBUser             string
	DBPassword         string
	DBName             string
	DBSSLMode          string
	DBTimeZone         string
	JWTSecret          string
	AdminBearerToken   string
	APNSTeamIdentifier string
	APNSKeyIdentifier  string
	APNSPrivateKeyPath string
	APNSEnvironment    string
	APNSBundleID       string
	AppMode            string
	Port               string
	SwaggerPassword    string
}

var AppConfig *Config

func InitConfig() {
	err := godotenv.Load()
	if err != nil {
		log.Println("[WARN] 未找到 .env 文件，将尝试直接使用系统环境变量")
	}

	AppConfig = &Config{
		DBHost:             getEnvOrFatal("DB_HOST"),
		DBPort:             getEnvOrFatal("DB_PORT"),
		DBUser:             getEnvOrFatal("DB_USER"),
		DBPassword:         getEnvOrFatal("DB_PASSWORD"),
		DBName:             getEnvOrFatal("DB_NAME"),
		DBSSLMode:          getEnvOrFatal("DB_SSLMODE"),
		DBTimeZone:         getEnvOrFatal("DB_TIMEZONE"),
		JWTSecret:          getEnvOrFatal("JWT_SECRET"),
		AdminBearerToken:   getEnvOrFatal("ADMIN_BEARER_TOKEN"),
		APNSTeamIdentifier: getEnvOrFatal("APNS_TEAM_IDENTIFIER"),
		APNSKeyIdentifier:  getEnvOrFatal("APNS_KEY_IDENTIFIER"),
		APNSPrivateKeyPath: getEnvOrFatal("APNS_PRIVATE_KEY_PATH"),
		APNSEnvironment:    getEnvOrFatal("APNS_ENVIRONMENT"),
		APNSBundleID:       getEnvOrFatal("APNS_BUNDLE_ID"),
		AppMode:            getEnvOrDefault("APP_MODE", "development"),
		Port:               getEnvOrDefault("PORT", "7241"),
		SwaggerPassword:    getEnvOrFatal("SWAGGER_PASSWORD"),
	}
}

func getEnvOrDefault(key, defaultValue string) string {
	val := os.Getenv(key)
	if val == "" {
		return defaultValue
	}
	return val
}

func getEnvOrFatal(key string) string {
	val := os.Getenv(key)
	if val == "" {
		log.Fatalf("[FATAL] 缺少必要的环境变量配置: %s", key)
	}
	return val
}
