package config

import (
	"log"
	"os"
	"strings"

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
	AdminBearerToken   string
	CORSAllowedOrigins []string
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
		AdminBearerToken:   getEnvOrFatal("ADMIN_BEARER_TOKEN"),
		CORSAllowedOrigins: parseCommaSeparatedEnv("CORS_ALLOWED_ORIGINS"),
		AppMode:            getEnvOrDefault("APP_MODE", "development"),
		Port:               getEnvOrDefault("PORT", "7241"),
		SwaggerPassword:    getEnvOrFatal("SWAGGER_PASSWORD"),
	}
}

func parseCommaSeparatedEnv(key string) []string {
	raw := os.Getenv(key)
	if raw == "" {
		return nil
	}

	seen := make(map[string]struct{})
	values := make([]string, 0)
	for _, item := range strings.Split(raw, ",") {
		value := strings.TrimSpace(item)
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		values = append(values, value)
	}

	return values
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
