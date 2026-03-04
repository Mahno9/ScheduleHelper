package handlers

import (
	"encoding/json"
	"net/http"
	"schedulehelper/db"
	"schedulehelper/models"

	"github.com/google/uuid"
)

func SlotsHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	switch r.Method {
	case http.MethodGet:
		rows, err := db.DB.Query("SELECT id, user_id, start_time, end_time, comment FROM slots WHERE user_id = ?", userID)
		if err != nil {
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var slots []models.Slot
		for rows.Next() {
			var s models.Slot
			if err := rows.Scan(&s.ID, &s.UserID, &s.StartTime, &s.EndTime, &s.Comment); err != nil {
				http.Error(w, "Database scan error", http.StatusInternalServerError)
				return
			}
			slots = append(slots, s)
		}
		if slots == nil {
			slots = []models.Slot{}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(slots)

	case http.MethodPost:
		var s models.Slot
		if err := json.NewDecoder(r.Body).Decode(&s); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		s.ID = uuid.New().String()
		s.UserID = userID

		_, err := db.DB.Exec("INSERT INTO slots (id, user_id, start_time, end_time, comment) VALUES (?, ?, ?, ?, ?)",
			s.ID, s.UserID, s.StartTime, s.EndTime, s.Comment)
		if err != nil {
			http.Error(w, "Failed to create slot", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(s)

	case http.MethodPut:
		var s models.Slot
		if err := json.NewDecoder(r.Body).Decode(&s); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}
		var owner string
		if err := db.DB.QueryRow("SELECT user_id FROM slots WHERE id = ?", s.ID).Scan(&owner); err != nil || owner != userID {
			http.Error(w, "Forbidden or not found", http.StatusForbidden)
			return
		}

		_, err := db.DB.Exec("UPDATE slots SET start_time = ?, end_time = ?, comment = ? WHERE id = ?",
			s.StartTime, s.EndTime, s.Comment, s.ID)
		if err != nil {
			http.Error(w, "Failed to update slot", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)

	case http.MethodDelete:
		slotID := r.URL.Query().Get("id")
		if slotID == "" {
			http.Error(w, "Missing slot ID", http.StatusBadRequest)
			return
		}

		_, err := db.DB.Exec("DELETE FROM slots WHERE id = ? AND user_id = ?", slotID, userID)
		if err != nil {
			http.Error(w, "Failed to delete slot", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}