package config

import "os"

type Config struct {
    HTTPPort     string
    GRPCPort     string
    DBHost       string
    DBPort       string
    DBUser       string
    DBPassword   string
    DBName       string
    RedisHost    string
    RedisPort    string
    MinioHost    string
    MinioApiPort string
    MinioUser    string
    MinioPass    string
    JWTSecret    string
}

func LoadConfig() *Config {
    return &Config{
        HTTPPort:     getEnv("USER_HTTP_PORT", "8081"),
        GRPCPort:     getEnv("USER_GRPC_PORT", "9091"),
        // Используем DB_USER и DB_PASSWORD из твоего .env
        DBHost:       getEnv("USER_DB_HOST", "user_postgres"),
        DBPort:       getEnv("DB_PORT", "5432"),
        DBUser:       getEnv("DB_USER", "chatuser"),
        DBPassword:   getEnv("DB_PASSWORD", "chatpass123"),
        DBName:       getEnv("USER_DB_NAME", "chatapp_user"),
        // Redis должен называться так же, как сервис в docker-compose
        RedisHost:    getEnv("REDIS_HOST", "redis"), 
        RedisPort:    getEnv("REDIS_PORT", "6379"),
        MinioHost:    getEnv("MINIO_HOST", "minio"),
        MinioApiPort: getEnv("MINIO_API_PORT", "9000"),
        MinioUser:    getEnv("MINIO_USER", "admin"),
        MinioPass:    getEnv("MINIO_PASSWORD", "admin123"),
        JWTSecret:    getEnv("JWT_SECRET", "your_secret_key"),
    }
}

func getEnv(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}