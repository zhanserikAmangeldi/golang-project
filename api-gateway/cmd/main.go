package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/websocket"
	"golang.org/x/time/rate"
)

var (
	userServiceURL = getEnv("AUTH_SERVICE_URL", "http://localhost:8081")
	chatServiceURL = getEnv("CHAT_SERVICE_URL", "http://localhost:8082")
	chatWSURL      = getEnv("CHAT_WS_URL", "ws://localhost:8082")
)

var jwtSecret []byte

var limiter = rate.NewLimiter(rate.Limit(100), 200)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func init() {
	secret := getEnv("JWT_SECRET", "")
	if secret == "" {
		log.Fatal("JWT_SECRET environment variable is required")
	}
	jwtSecret = []byte(secret)
}

func main() {
	if getEnv("GIN_MODE", "debug") == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()

	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.Use(rateLimitMiddleware())

	allowedOrigins := strings.Split(getEnv("ALLOWED_ORIGINS", "*"), ",")
	r.Use(cors.New(cors.Config{
		AllowOrigins:     allowedOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Authorization", "Content-Type", "X-User-ID"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	r.GET("/health", healthCheck)
	r.GET("/readiness", readinessCheck)

	r.GET("/ws", handleWebSocket)

	api := r.Group("/api/v1")
	{
		user := api.Group("/users")
		{
			auth := user.Group("/auth")
			{
				auth.POST("/register", proxyRequest(userServiceURL, 10*time.Second))
				auth.POST("/login", proxyRequest(userServiceURL, 10*time.Second))
				auth.POST("/refresh", proxyRequest(userServiceURL, 5*time.Second))
				auth.POST("/logout", proxyRequest(userServiceURL, 5*time.Second))
			}

			user.Use(authMiddleware())
			{
				auth := user.Group("/auth")
				{
					auth.POST("/logout-all", proxyRequest(userServiceURL, 5*time.Second))
					auth.GET("/sessions", proxyRequest(userServiceURL, 5*time.Second))
				}

				user.POST("/upload-avatar", proxyRequest(userServiceURL, 30*time.Second))
				user.GET("/get-avatar", proxyRequest(userServiceURL, 10*time.Second))
				user.GET("/me", proxyRequest(userServiceURL, 5*time.Second))
				user.PUT("/me", proxyRequest(userServiceURL, 10*time.Second))
				user.GET("/:id", proxyRequest(userServiceURL, 5*time.Second))

				admin := user.Group("/admin")
				{
					admin.GET("/users", proxyRequest(userServiceURL, 10*time.Second))
					admin.DELETE("/users/:user_id", proxyRequest(userServiceURL, 10*time.Second))
					admin.PUT("/users/role", proxyRequest(userServiceURL, 10*time.Second))
					admin.POST("/bans", proxyRequest(userServiceURL, 10*time.Second))
					admin.DELETE("/bans/:user_id", proxyRequest(userServiceURL, 10*time.Second))
					admin.GET("/bans/:user_id/history", proxyRequest(userServiceURL, 10*time.Second))
					admin.GET("/audit-logs", proxyRequest(userServiceURL, 10*time.Second))
					admin.GET("/audit-logs/user/:user_id", proxyRequest(userServiceURL, 10*time.Second))
					admin.GET("/stats", proxyRequest(userServiceURL, 10*time.Second))
				}

				moderator := user.Group("/moderator")
				{
					moderator.GET("/users", proxyRequest(userServiceURL, 10*time.Second))
					moderator.GET("/bans/:user_id/history", proxyRequest(userServiceURL, 10*time.Second))
				}
			}
		}

		chat := api.Group("/chat")
		chat.Use(authMiddleware())
		{
			groups := chat.Group("/groups")
			{
				groups.POST("/create", proxyRequest(chatServiceURL, 10*time.Second))
			}

			messages := chat.Group("/messages")
			{
				messages.POST("/send", proxyRequest(chatServiceURL, 10*time.Second))
				messages.GET("/history", proxyRequest(chatServiceURL, 10*time.Second))
				messages.POST("/read", proxyRequest(chatServiceURL, 5*time.Second))
				messages.POST("/edit", proxyRequest(chatServiceURL, 10*time.Second))
				messages.DELETE("/delete", proxyRequest(chatServiceURL, 10*time.Second))

				reactions := messages.Group("/reactions")
				{
					reactions.POST("/add", proxyRequest(chatServiceURL, 5*time.Second))
					reactions.DELETE("/remove", proxyRequest(chatServiceURL, 5*time.Second))
				}
			}

			chat.GET("/conversations", proxyRequest(chatServiceURL, 10*time.Second))

			files := chat.Group("/files")
			{
				files.POST("/upload", proxyRequest(chatServiceURL, 60*time.Second))
				files.POST("/send", proxyRequest(chatServiceURL, 30*time.Second))
				files.GET("/get", proxyRequest(chatServiceURL, 30*time.Second))
			}
		}
	}

	r.GET("/verify-email", proxyRequest(userServiceURL, 10*time.Second))

	port := getEnv("PORT", "8000")
	log.Printf("API Gateway starting on port %s", port)
	log.Printf("User Service: %s", userServiceURL)
	log.Printf("Chat Service: %s", chatServiceURL)
	log.Printf("Chat WebSocket: %s", chatWSURL)

	if err := r.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}

func rateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !limiter.Allow() {
			log.Printf("Rate limit exceeded for IP: %s", c.ClientIP())
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":   "Rate limit exceeded",
				"message": "Too many requests. Please try again later.",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}

func authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")

		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Authorization required",
				"message": "No authorization header provided",
			})
			c.Abort()
			return
		}

		if !strings.HasPrefix(authHeader, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Invalid authorization format",
				"message": "Authorization header must start with 'Bearer '",
			})
			c.Abort()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return jwtSecret, nil
		})

		if err != nil {
			log.Printf("JWT parse error: %v", err)
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Invalid token",
				"message": err.Error(),
			})
			c.Abort()
			return
		}

		if !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Token invalid",
				"message": "Token validation failed",
			})
			c.Abort()
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Could not parse token claims",
			})
			c.Abort()
			return
		}

		var userID string
		if uid, ok := claims["user_id"]; ok {
			switch v := uid.(type) {
			case float64:
				userID = strconv.Itoa(int(v))
			case string:
				userID = v
			case int:
				userID = strconv.Itoa(v)
			default:
				log.Printf("Unexpected user_id type: %T", v)
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "Invalid user_id format in token",
				})
				c.Abort()
				return
			}
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Token missing user_id claim",
			})
			c.Abort()
			return
		}

		c.Set("user_id", userID)

		if email, ok := claims["email"].(string); ok {
			c.Set("user_email", email)
		}
		if username, ok := claims["username"].(string); ok {
			c.Set("user_username", username)
		}

		c.Next()
	}
}

func proxyRequest(targetURL string, timeout time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		target, err := url.Parse(targetURL)
		if err != nil {
			log.Printf("Failed to parse target URL %s: %v", targetURL, err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Invalid service configuration",
			})
			return
		}

		proxy := &httputil.ReverseProxy{
			Director: func(req *http.Request) {
				req.URL.Scheme = target.Scheme
				req.URL.Host = target.Host
				req.Host = target.Host

				originalPath := c.Request.URL.Path
				targetPath := buildTargetPath(originalPath)

				if c.Request.URL.RawQuery != "" {
					targetPath += "?" + c.Request.URL.RawQuery
				}

				req.URL.Path = targetPath

				if userID, exists := c.Get("user_id"); exists {
					req.Header.Set("X-User-ID", userID.(string))
				}
				if email, exists := c.Get("user_email"); exists {
					req.Header.Set("X-User-Email", email.(string))
				}
				if username, exists := c.Get("user_username"); exists {
					req.Header.Set("X-User-Username", username.(string))
				}

				req.Header.Set("X-Forwarded-By", "API-Gateway")
				req.Header.Set("X-Forwarded-For", c.ClientIP())
				req.Header.Set("X-Real-IP", c.ClientIP())
				req.Header.Set("X-Forwarded-Proto", c.Request.Proto)

				if gin.Mode() == gin.DebugMode {
					log.Printf("Proxying: %s %s → %s%s",
						req.Method, originalPath, target.Host, targetPath)
				}
			},
			ErrorHandler: func(rw http.ResponseWriter, req *http.Request, err error) {
				log.Printf("Proxy error for %s: %v", target.Host, err)

				c.JSON(http.StatusBadGateway, gin.H{
					"error":   "Service temporarily unavailable",
					"service": target.Host,
					"details": err.Error(),
				})
			},
		}

		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
		defer cancel()

		c.Request = c.Request.WithContext(ctx)
		proxy.ServeHTTP(c.Writer, c.Request)
	}
}

func buildTargetPath(originalPath string) string {
	if strings.HasPrefix(originalPath, "/api/v1/users") {
		path := strings.TrimPrefix(originalPath, "/api/v1/users")
		if path == "" {
			return "/api/v1/users"
		}
		return "/api/v1" + path
	}

	if strings.HasPrefix(originalPath, "/api/v1/chat") {
		path := strings.TrimPrefix(originalPath, "/api/v1/chat")
		if path == "" {
			return "/api/v1"
		}
		return "/api/v1" + path
	}

	return originalPath
}

func handleWebSocket(c *gin.Context) {
	tokenString := c.Query("token")
	if tokenString == "" {
		authHeader := c.GetHeader("Authorization")
		if strings.HasPrefix(authHeader, "Bearer ") {
			tokenString = strings.TrimPrefix(authHeader, "Bearer ")
		}
	}

	if tokenString == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "No authentication token provided",
		})
		return
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return jwtSecret, nil
	})

	if err != nil || !token.Valid {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Invalid token",
		})
		return
	}

	clientConn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}
	defer clientConn.Close()

	backendURL := strings.Replace(chatWSURL, "http://", "ws://", 1)
	backendURL = strings.Replace(backendURL, "https://", "wss://", 1)
	backendURL += "/ws?token=" + tokenString

	backendConn, _, err := websocket.DefaultDialer.Dial(backendURL, nil)
	if err != nil {
		log.Printf("Failed to connect to backend WebSocket: %v", err)
		clientConn.WriteMessage(websocket.CloseMessage,
			websocket.FormatCloseMessage(websocket.CloseInternalServerErr, "Service unavailable"))
		return
	}
	defer backendConn.Close()

	log.Printf("WebSocket connection established for token")

	done := make(chan struct{})

	go func() {
		defer close(done)
		for {
			messageType, message, err := clientConn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					log.Printf("WebSocket read error: %v", err)
				}
				return
			}
			if err := backendConn.WriteMessage(messageType, message); err != nil {
				log.Printf("WebSocket write error to backend: %v", err)
				return
			}
		}
	}()

	go func() {
		for {
			messageType, message, err := backendConn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					log.Printf("WebSocket read error from backend: %v", err)
				}
				return
			}
			if err := clientConn.WriteMessage(messageType, message); err != nil {
				log.Printf("WebSocket write error to client: %v", err)
				return
			}
		}
	}()

	<-done
	log.Printf("WebSocket connection closed")
}

func healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"timestamp": time.Now().Unix(),
		"service":   "api-gateway",
		"version":   "1.0.0",
	})
}

func readinessCheck(c *gin.Context) {
	services := map[string]string{
		"user": userServiceURL + "/health",
		"chat": chatServiceURL + "/health",
	}

	results := make(map[string]map[string]interface{})
	allReady := true

	for name, url := range services {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
		start := time.Now()
		resp, err := http.DefaultClient.Do(req)
		latency := time.Since(start).Milliseconds()

		ready := err == nil && resp != nil && resp.StatusCode == http.StatusOK

		results[name] = map[string]interface{}{
			"ready":   ready,
			"latency": latency,
		}

		if !ready {
			allReady = false
			if err != nil {
				results[name]["error"] = err.Error()
			}
		}

		if resp != nil {
			resp.Body.Close()
		}
	}

	status := http.StatusOK
	if !allReady {
		status = http.StatusServiceUnavailable
	}

	c.JSON(status, gin.H{
		"ready":     allReady,
		"services":  results,
		"timestamp": time.Now().Unix(),
	})
}
