package handlers

import (
	"encoding/json"
	"net/http"
	"schedulehelper/db"
	"schedulehelper/models"

	"github.com/google/uuid"
)

type EventRequest struct {
	models.Event
	Participants []string `json:"participants"`
}

func EventsHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	switch r.Method {
	case http.MethodGet:
		rows, err := db.DB.Query("SELECT id, title, description, icon, color, start_time, end_time FROM events")
		if err != nil {
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var events []EventRequest
		for rows.Next() {
			var e EventRequest
			if err := rows.Scan(&e.ID, &e.Title, &e.Description, &e.Icon, &e.Color, &e.StartTime, &e.EndTime); err != nil {
				continue
			}
			
			pRows, _ := db.DB.Query("SELECT user_id FROM event_participants WHERE event_id = ?", e.ID)
			for pRows.Next() {
				var p string
				pRows.Scan(&p)
				e.Participants = append(e.Participants, p)
			}
			pRows.Close()
			if e.Participants == nil {
				e.Participants = []string{}
			}

			events = append(events, e)
		}
		if events == nil {
			events = []EventRequest{}
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(events)

	case http.MethodPost:
		var req EventRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}
		req.ID = uuid.New().String()
		
		tx, err := db.DB.Begin()
		if err != nil {
			http.Error(w, "Tx error", http.StatusInternalServerError)
			return
		}

		_, err = tx.Exec("INSERT INTO events (id, title, description, icon, color, start_time, end_time) VALUES (?, ?, ?, ?, ?, ?, ?)",
			req.ID, req.Title, req.Description, req.Icon, req.Color, req.StartTime, req.EndTime)
		if err != nil {
			tx.Rollback()
			http.Error(w, "Failed to create event", http.StatusInternalServerError)
			return
		}

		for _, pID := range req.Participants {
			_, err = tx.Exec("INSERT INTO event_participants (event_id, user_id) VALUES (?, ?)", req.ID, pID)
			if err != nil {
				tx.Rollback()
				http.Error(w, "Failed to add participant", http.StatusInternalServerError)
				return
			}
		}
		tx.Commit()
		
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(req)

	case http.MethodPut:
		var req EventRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}
		
		tx, err := db.DB.Begin()
		if err != nil {
			http.Error(w, "Tx error", http.StatusInternalServerError)
			return
		}

		_, err = tx.Exec("UPDATE events SET title=?, description=?, icon=?, color=?, start_time=?, end_time=? WHERE id=?",
			req.Title, req.Description, req.Icon, req.Color, req.StartTime, req.EndTime, req.ID)
		if err != nil {
			tx.Rollback()
			http.Error(w, "Failed to update event", http.StatusInternalServerError)
			return
		}

		tx.Exec("DELETE FROM event_participants WHERE event_id=?", req.ID)
		for _, pID := range req.Participants {
			tx.Exec("INSERT INTO event_participants (event_id, user_id) VALUES (?, ?)", req.ID, pID)
		}
		tx.Commit()
		w.WriteHeader(http.StatusOK)

	case http.MethodDelete:
		eventID := r.URL.Query().Get("id")
		if eventID == "" {
			http.Error(w, "Missing id", http.StatusBadRequest)
			return
		}
		db.DB.Exec("DELETE FROM events WHERE id=?", eventID)
		w.WriteHeader(http.StatusNoContent)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}