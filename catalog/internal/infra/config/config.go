package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	HttpListener string
	GrpcListener string
	Db           DbConfig
}

type DbConfig struct {
	User     string
	Password string
	Address  string
	Name     string
}

func NewConfig() *Config {
	if err := godotenv.Load(); err != nil {
		fmt.Printf("warning: .env file not found: %v\n", err)
	}
	return &Config{
		HttpListener: getEnv("HTTP_PORT", ":8080"),
		GrpcListener: getEnv("GRPC_PORT", "9090"),
		Db: DbConfig{
			User:     getEnv("DB_USER", "root"),
			Password: getEnv("DB_PASS", "root"),
			Address:  fmt.Sprintf("%s:%s", getEnv("DB_HOST", "127.0.0.1"), getEnv("DB_PORT", "3306")),
			Name:     getEnv("DB_NAME", "catalog-db"),
		},
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if value, ok := os.LookupEnv(key); ok {
		var i int
		if _, err := fmt.Sscanf(value, "%d", &i); err == nil {
			return i
		}
	}
	return fallback
}
