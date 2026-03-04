package handlers

import (
	"encoding/json"
	"net/http"
	"schedulehelper/db"
	"schedulehelper/models"
)

func ProfileHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	switch r.Method {
	case http.MethodGet:
		var u models.User
		err := db.DB.QueryRow("SELECT id, username, color, emoji, theme, timezone, created_at FROM users WHERE id = ?", userID).
			Scan(&u.ID, &u.Username, &u.Color, &u.Emoji, &u.Theme, &u.Timezone, &u.CreatedAt)
		if err != nil {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(u)

	case http.MethodPut:
		var u models.User
		if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}
		
		if u.Username != "" {
			var exists bool
			db.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE username = ? AND id != ?)", u.Username, userID).Scan(&exists)
			if exists {
				http.Error(w, "Username already taken", http.StatusConflict)
				return
			}
		}

		_, err := db.DB.Exec("UPDATE users SET username = ?, color = ?, emoji = ?, theme = ?, timezone = ? WHERE id = ?",
			u.Username, u.Color, u.Emoji, u.Theme, u.Timezone, userID)
		if err != nil {
			http.Error(w, "Failed to update profile", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)

	case http.MethodDelete:
		_, err := db.DB.Exec("DELETE FROM users WHERE id = ?", userID)
		if err != nil {
			http.Error(w, "Failed to delete profile", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}