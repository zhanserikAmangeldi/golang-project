package main

import (
	"log"
	"net/http"

	"github.com/zhanserikAmangeldi/chat-service/internal/api"
	"github.com/zhanserikAmangeldi/chat-service/internal/config"
	"github.com/zhanserikAmangeldi/chat-service/internal/db"
)

func main() {
	cfg := config.Load()

	database := db.Connect(cfg.DBUrl)
	defer database.Close()

	router := api.NewRouter(database, cfg)

	log.Printf("chat-service starting on %s", cfg.HTTPPort)
	if err := http.ListenAndServe(cfg.HTTPPort, router); err != nil {
		log.Fatal(err)
	}
}
