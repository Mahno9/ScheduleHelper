package handlers

import (
	"encoding/json"
	"net/http"
	"schedulehelper/db"
	"schedulehelper/models"
)

func GetUsersHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	rows, err := db.DB.Query("SELECT id, username, color, emoji, theme, timezone, created_at FROM users")
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var u models.User
		if err := rows.Scan(&u.ID, &u.Username, &u.Color, &u.Emoji, &u.Theme, &u.Timezone, &u.CreatedAt); err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		users = append(users, u)
	}

	if users == nil {
		users = []models.User{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}