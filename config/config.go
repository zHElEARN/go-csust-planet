package config

import (
	"fmt"
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

func Load() (Config, error) {
	err := godotenv.Load()
	if err != nil {
		log.Println("[WARN] 未找到 .env 文件，将尝试直接使用系统环境变量")
	}

	cfg := Config{
		DBHost:             os.Getenv("DB_HOST"),
		DBPort:             os.Getenv("DB_PORT"),
		DBUser:             os.Getenv("DB_USER"),
		DBPassword:         os.Getenv("DB_PASSWORD"),
		DBName:             os.Getenv("DB_NAME"),
		DBSSLMode:          os.Getenv("DB_SSLMODE"),
		DBTimeZone:         os.Getenv("DB_TIMEZONE"),
		AdminBearerToken:   os.Getenv("ADMIN_BEARER_TOKEN"),
		CORSAllowedOrigins: parseCommaSeparatedEnv("CORS_ALLOWED_ORIGINS"),
		AppMode:            getEnvOrDefault("APP_MODE", "development"),
		Port:               getEnvOrDefault("PORT", "7241"),
		SwaggerPassword:    os.Getenv("SWAGGER_PASSWORD"),
	}

	for key, value := range map[string]string{
		"DB_HOST": cfg.DBHost, "DB_PORT": cfg.DBPort, "DB_USER": cfg.DBUser,
		"DB_PASSWORD": cfg.DBPassword, "DB_NAME": cfg.DBName, "DB_SSLMODE": cfg.DBSSLMode,
		"DB_TIMEZONE": cfg.DBTimeZone, "ADMIN_BEARER_TOKEN": cfg.AdminBearerToken,
		"SWAGGER_PASSWORD": cfg.SwaggerPassword,
	} {
		if value == "" {
			return Config{}, fmt.Errorf("缺少必要的环境变量配置: %s", key)
		}
	}

	return cfg, nil
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
