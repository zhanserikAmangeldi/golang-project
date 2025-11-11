package config

import "os"

type Config struct {
	HTTPPort       string
	GRPCPort       string
	DBUrl          string
	DBPort         string
	DBUser         string
	DBPassword     string
	DBName         string
	RedisHost      string
	RedisPort      string
	UserServiceURL string
	JWTSecret      string
}

func Load() Config {
	return Config{
		HTTPPort:       getEnv("PORT", ":8082"),
		DBUrl:          getEnv("DB_URL", "postgres://postgres:password@localhost:5432/chat?sslmode=disable"),
		UserServiceURL: getEnv("USER_SERVICE_URL", "http://localhost:8081"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
