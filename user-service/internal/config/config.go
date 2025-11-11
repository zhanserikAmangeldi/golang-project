package config

import "os"

type Config struct {
	HTTPPort   string
	GRPCPort   string
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	RedisHost  string
	RedisPort  string
	JWTSecret  string
}

func LoadConfig() *Config {
	return &Config{
		HTTPPort:   getEnv("HTTP_PORT", "8081"),
		GRPCPort:   getEnv("GRPC_PORT", "9091"),
		DBHost:     getEnv("DB_HOST", "chat_postgres"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBUser:     getEnv("DB_USER", "chatuser"),
		DBPassword: getEnv("DB_PASSWORD", "chatpass123"),
		DBName:     getEnv("DB_NAME", "chatapp"),
		RedisHost:  getEnv("REDIS_HOST", "chat_redis"),
		RedisPort:  getEnv("REDIS_PORT", "6379"),
		JWTSecret:  getEnv("JWT_SECRET", "your-super-secret-key"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
