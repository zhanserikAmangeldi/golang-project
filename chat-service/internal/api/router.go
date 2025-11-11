package api

import (
	"database/sql"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/zhanserikAmangeldi/chat-service/internal/config"
	"github.com/zhanserikAmangeldi/chat-service/internal/handlers"
	"github.com/zhanserikAmangeldi/chat-service/internal/middleware"
)

func NewRouter(db *sql.DB, cfg config.Config) http.Handler {
	r := mux.NewRouter()

	r.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	}).Methods("GET")

	protected := r.PathPrefix("/").Subrouter()
	protected.Use(middleware.AuthMiddleware)

	protected.HandleFunc("/conversations/{id}/messages", handlers.SendMessage(db, cfg)).Methods("POST")
	protected.HandleFunc("/conversations/{id}/messages", handlers.GetMessages(db)).Methods("GET")

	return r
}
