package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	HTTPPort       string
	GRPCPort       string
	DBHost         string
	DBPort         string
	DBUser         string
	DBPassword     string
	DBName         string
	DBSSLMode      string
	RedisHost      string
	RedisPort      string
	RedisDB        int
	UserServiceURL string
	JWTSecret      string
}

func Load() *Config {
	redisDB, _ := strconv.Atoi(getEnv("REDIS_DB", "0"))

	return &Config{
		HTTPPort:       getEnv("HTTP_PORT", "8082"),
		GRPCPort:       getEnv("GRPC_PORT", "9092"),
		DBHost:         getEnv("CHAT_DB_HOST", "localhost"),
		DBPort:         getEnv("CHAT_DB_PORT", "5432"),
		DBUser:         getEnv("CHAT_DB_USER", "postgres"),
		DBPassword:     getEnv("CHAT_DB_PASSWORD", "secret"),
		DBName:         getEnv("CHAT_DB_NAME", "postgres"),
		DBSSLMode:      getEnv("CHAT_DB_SSLMODE", "disable"),
		RedisHost:      getEnv("REDIS_HOST", "localhost"),
		RedisPort:      getEnv("REDIS_PORT", "6379"),
		RedisDB:        redisDB,
		UserServiceURL: getEnv("USER_SERVICE_URL", "http://user_service:8081"),
		JWTSecret:      getEnv("JWT_SECRET", "your-super-secret-key"),
	}
}

func (c *Config) GetDBConnectionString() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.DBHost, c.DBPort, c.DBUser, c.DBPassword, c.DBName, c.DBSSLMode,
	)
}

func (c *Config) GetDBURL() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		c.DBUser, c.DBPassword, c.DBHost, c.DBPort, c.DBName, c.DBSSLMode,
	)
}

func (c *Config) GetRedisAddr() string {
	return fmt.Sprintf("%s:%s", c.RedisHost, c.RedisPort)
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
