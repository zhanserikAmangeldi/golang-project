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
    // Поля для MinIO
    MinioHost      string
    MinioApiPort   string
    MinioAccessKey string
    MinioSecretKey string
    MinioUseSSL    bool
}

func Load() *Config {
    redisDB, _ := strconv.Atoi(getEnv("REDIS_DB", "0"))
    useSSL, _ := strconv.ParseBool(getEnv("MINIO_USE_SSL", "false"))

    return &Config{
        HTTPPort:       getEnv("CHAT_HTTP_PORT", "8082"),
        GRPCPort:       getEnv("CHAT_GRPC_PORT", "9092"),
        DBHost:         getEnv("CHAT_DB_HOST", "chat_postgres"),
        DBPort:         getEnv("DB_PORT", "5432"),
        DBUser:         getEnv("DB_USER", "chatuser"),
        DBPassword:     getEnv("DB_PASSWORD", "chatpass123"),
        DBName:         getEnv("CHAT_DB_NAME", "chatapp_chat"),
        DBSSLMode:      getEnv("DB_SSLMODE", "disable"),
        RedisHost:      getEnv("REDIS_HOST", "redis"),
        RedisPort:      getEnv("REDIS_PORT", "6379"),
        RedisDB:        redisDB,
        UserServiceURL: getEnv("USER_SERVICE_URL", "user_service:9091"),
        JWTSecret:      getEnv("JWT_SECRET", "your_secret_key"),
        MinioHost:      getEnv("MINIO_HOST", "minio"),
        MinioApiPort:   getEnv("MINIO_API_PORT", "9000"),
        MinioAccessKey: getEnv("MINIO_USER", "admin"),
        MinioSecretKey: getEnv("MINIO_PASSWORD", "admin123"),
        MinioUseSSL:    useSSL,
    }
}

// Используется для sqlx.Connect
func (c *Config) GetDBConnectionString() string {
    return fmt.Sprintf(
        "host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
        c.DBHost, c.DBPort, c.DBUser, c.DBPassword, c.DBName, c.DBSSLMode,
    )
}

// Используется для миграций (то, что просил main.go)
func (c *Config) GetDBURL() string {
    return fmt.Sprintf(
        "postgres://%s:%s@%s:%s/%s?sslmode=%s",
        c.DBUser, c.DBPassword, c.DBHost, c.DBPort, c.DBName, c.DBSSLMode,
    )
}

func getEnv(key, fallback string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return fallback
}