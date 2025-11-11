package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"github.com/zhanserikAmangeldi/user-service/internal/config"
	"github.com/zhanserikAmangeldi/user-service/internal/handler"
	"github.com/zhanserikAmangeldi/user-service/internal/mailer"
	"github.com/zhanserikAmangeldi/user-service/internal/middleware"
	"github.com/zhanserikAmangeldi/user-service/internal/repository"
	"github.com/zhanserikAmangeldi/user-service/internal/service"
	"github.com/zhanserikAmangeldi/user-service/pkg/jwt"
)

func main() {
	cfg := config.LoadConfig()
	ctx := context.Background()

	//dbURL := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
	//	cfg.DBUser, cfg.DBPassword, cfg.DBHost, cfg.DBPort, cfg.DBName)
	dbURL := "postgres://chatuser:chatpass123@chat_postgres:5432/chatapp?sslmode=disable"

	dbPool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v", err)
	}
	defer dbPool.Close()

	if err := dbPool.Ping(ctx); err != nil {
		log.Fatalf("Unable to ping database: %v", err)
	}
	log.Println("Connected to PostgreSQL")

	redisClient := redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%s", cfg.RedisHost, cfg.RedisPort),
		DB:   0,
	})
	defer redisClient.Close()

	if err := redisClient.Ping(ctx).Err(); err != nil {
		log.Fatalf("Unable to connect to Redis: %v", err)
	}
	log.Println("Connected to Redis")

	render := mailer.NewTemplateRender("internal/mailer/templates")

	smtp := mailer.SMTPMailer{
		Host:    "smtp.gmail.com",
		Port:    587,
		User:    "amangeldi.janserik2017@gmail.com",
		Pass:    "xsts bhls apvn ucol",
		From:    "Your new best chat application :))) <noreply@chat.com>",
		BaseURL: "localhost:8081",
		Render:  render,
	}

	userRepo := repository.NewUserRepository(dbPool)
	sessionRepo := repository.NewSessionRepository(dbPool)
	emailRepo := repository.NewEmailVerificationRepository(dbPool)

	tokenManager := jwt.NewTokenManager(cfg.JWTSecret)
	authService := service.NewAuthService(userRepo, sessionRepo, tokenManager, emailRepo, &smtp, redisClient)

	authHandler := handler.NewAuthHandler(authService)
	userHandler := handler.NewUserHandler(userRepo)
	emailVerificationHandler := handler.NewEmailVerificationHandler(authService)

	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()

	router.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":   "healthy",
			"service":  "user-service",
			"database": "connected",
			"redis":    "connected",
		})
	})

	router.GET("/verify-email", emailVerificationHandler.VerifyEmail)

	v1 := router.Group("/api/v1")
	{
		auth := v1.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.POST("/refresh", authHandler.RefreshToken)
			auth.POST("/logout", authHandler.Logout)
		}
	}

	protected := v1.Group("")
	protected.Use(middleware.AuthMiddleware(tokenManager, redisClient))
	{
		auth := protected.Group("/auth")
		{
			auth.POST("/logout-all", authHandler.LogoutAll)
			auth.GET("/sessions", authHandler.GetActiveSessions)
		}

		users := protected.Group("/users")
		{
			users.GET("/me", userHandler.GetMe)
			users.PUT("/me", userHandler.UpdateMe)
			users.GET("/:id", userHandler.GetUserByID)
		}
	}

	srv := &http.Server{
		Addr:    ":" + cfg.HTTPPort,
		Handler: router,
	}

	log.Printf("User service starting on port %s", cfg.HTTPPort)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Failed to start server: %v", err)
	}
}
