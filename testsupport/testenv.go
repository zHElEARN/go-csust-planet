package testsupport

import (
	"os"
	"path/filepath"
	"runtime"
	"sync"

	"github.com/joho/godotenv"
)

type DBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
	TimeZone string
}

var loadEnvOnce sync.Once

func LoadTestDBConfig() (DBConfig, bool) {
	loadEnvOnce.Do(func() {
		_ = godotenv.Load(filepath.Join(repoRoot(), ".env"))
	})

	host := os.Getenv("TEST_DB_HOST")
	port := os.Getenv("TEST_DB_PORT")
	user := os.Getenv("TEST_DB_USER")
	password := os.Getenv("TEST_DB_PASSWORD")
	name := os.Getenv("TEST_DB_NAME")
	if host == "" || port == "" || user == "" || password == "" || name == "" {
		return DBConfig{}, false
	}

	sslMode := os.Getenv("TEST_DB_SSLMODE")
	if sslMode == "" {
		sslMode = "disable"
	}

	timeZone := os.Getenv("TEST_DB_TIMEZONE")
	if timeZone == "" {
		timeZone = "Asia/Shanghai"
	}

	return DBConfig{
		Host:     host,
		Port:     port,
		User:     user,
		Password: password,
		Name:     name,
		SSLMode:  sslMode,
		TimeZone: timeZone,
	}, true
}

func repoRoot() string {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		return "."
	}

	return filepath.Clean(filepath.Join(filepath.Dir(filename), ".."))
}
