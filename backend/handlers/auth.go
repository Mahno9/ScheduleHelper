package handlers

import (
	"encoding/json"
	"net/http"
	"schedulehelper/db"
	"schedulehelper/models"
	"time"

	"github.com/google/uuid"
)

type RegisterRequest struct {
	Username string `json:"username"`
	Color    string `json:"color"`
	Emoji    string `json:"emoji"`
}

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Username == "" || req.Color == "" || req.Emoji == "" {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	var exists bool
	err := db.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE username = ?)", req.Username).Scan(&exists)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	if exists {
		http.Error(w, "Username already taken", http.StatusConflict)
		return
	}

	user := models.User{
		ID:        uuid.New().String(),
		Username:  req.Username,
		Color:     req.Color,
		Emoji:     req.Emoji,
		Theme:     "system",
		Timezone:  "auto",
		CreatedAt: time.Now(),
	}

	_, err = db.DB.Exec("INSERT INTO users (id, username, color, emoji, theme, timezone, created_at) VALUES (?, ?, ?, ?, ?, ?, ?)",
		user.ID, user.Username, user.Color, user.Emoji, user.Theme, user.Timezone, user.CreatedAt)
	if err != nil {
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}

type LoginRequest struct {
	ID string `json:"id"`
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	var u models.User
	err := db.DB.QueryRow("SELECT id, username, color, emoji, theme, timezone, created_at FROM users WHERE id = ?", req.ID).
		Scan(&u.ID, &u.Username, &u.Color, &u.Emoji, &u.Theme, &u.Timezone, &u.CreatedAt)
	if err != nil {
		http.Error(w, "User not found", http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(u)
}