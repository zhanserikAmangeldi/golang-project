package api

import (
	"database/sql"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/zhanserikAmangeldi/chat-service/config"
	"github.com/zhanserikAmangeldi/chat-service/internal/adapters/handler"
	"github.com/zhanserikAmangeldi/chat-service/internal/middleware"
)

func NewRouter(db *sql.DB, cfg *config.Config) http.Handler {
	r := mux.NewRouter()

	r.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	}).Methods("GET")

	protected := r.PathPrefix("/api/v1").Subrouter()

	protected.Use(middleware.AuthMiddleware)

	
	protected.HandleFunc("/conversations/{id}/messages", handler.SendMessage(db, cfg)).Methods("POST")
	protected.HandleFunc("/conversations/{id}/messages", handler.GetMessages(db)).Methods("GET")

	return r
}