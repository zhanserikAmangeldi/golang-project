package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/zhanserikAmangeldi/chat-service/internal/client"
	"github.com/zhanserikAmangeldi/chat-service/internal/config"
	"github.com/zhanserikAmangeldi/chat-service/internal/repository"
	"github.com/zhanserikAmangeldi/chat-service/internal/service"
)

func SendMessage(db *sql.DB, cfg config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		convID := vars["id"]

		var input struct {
			SenderID string `json:"sender_id"`
			Content  string `json:"content"`
			Type     string `json:"type"`
		}
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			http.Error(w, "invalid body", http.StatusBadRequest)
			return
		}

		// validate user exists via user-service
		uc := &client.UserClient{BaseURL: cfg.UserServiceURL}
		ok, err := uc.UserExists(input.SenderID)
		if err != nil {
			http.Error(w, "user validation failed", http.StatusInternalServerError)
			return
		}
		if !ok {
			http.Error(w, "user not found", http.StatusBadRequest)
			return
		}

		msgService := service.NewMessageService(repository.NewMessageRepo(db))
		msg, err := msgService.CreateMessage(convID, input.SenderID, input.Content, input.Type)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(msg)
	}
}

func GetMessages(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		convID := vars["id"]

		// pagination params: ?limit=20&before=<timestamp>
		q := r.URL.Query()
		limitStr := q.Get("limit")
		limit := 50
		if limitStr != "" {
			if v, err := strconv.Atoi(limitStr); err == nil && v > 0 && v <= 200 {
				limit = v
			}
		}
		before := q.Get("before") // RFC3339 timestamp
		var beforeTime time.Time
		if before != "" {
			t, err := time.Parse(time.RFC3339, before)
			if err == nil {
				beforeTime = t
			}
		}

		repo := repository.NewMessageRepo(db)
		msgs, err := repo.ListByConversation(convID, limit, beforeTime)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(msgs)
	}
}
