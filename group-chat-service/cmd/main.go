package main

import (
	"github.com/gin-gonic/gin"
	"github.com/zhanserikAmangeldi/group-chat-service/internal/handler"
	"github.com/zhanserikAmangeldi/group-chat-service/internal/websocket"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// 1. Setup DB
	dsn := "host=localhost user=postgres password=password dbname=chat port=5432 sslmode=disable"
	db, _ := gorm.Open(postgres.Open(dsn), &gorm.Config{})

	// 2. Setup WebSocket Hub
	hub := websocket.NewHub()
	go hub.Run()

	// 3. Setup Handlers
	groupHandler := handler.NewGroupHandler(db, hub)

	// 4. Setup Router
	r := gin.Default()

	v1 := r.Group("/api/v1")
	{
		v1.POST("/groups", groupHandler.CreateGroup)
		v1.POST("/groups/:id/join", groupHandler.JoinGroup)
		v1.GET("/ws/groups/:id", groupHandler.ConnectGroupChat)
	}

	r.Run(":8080")
}
