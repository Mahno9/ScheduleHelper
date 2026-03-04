package handlers

import (
	"encoding/json"
	"net/http"
	"schedulehelper/db"
	"schedulehelper/models"
	"time"
)

type CalendarData struct {
	Users  []models.User  `json:"users"`
	Slots  []models.Slot  `json:"slots"`
	Events []EventRequest `json:"events"`
}

func GetCalendarDataHandler(w http.ResponseWriter, r *http.Request) {
	startStr := r.URL.Query().Get("start")
	endStr := r.URL.Query().Get("end")

	var startDate, endDate time.Time
	if startStr != "" {
		startDate, _ = time.Parse(time.RFC3339, startStr)
	}
	if endStr != "" {
		endDate, _ = time.Parse(time.RFC3339, endStr)
	}
	
	var data CalendarData
	
	uRows, _ := db.DB.Query("SELECT id, username, color, emoji, theme, timezone, created_at FROM users")
	for uRows.Next() {
		var u models.User
		uRows.Scan(&u.ID, &u.Username, &u.Color, &u.Emoji, &u.Theme, &u.Timezone, &u.CreatedAt)
		data.Users = append(data.Users, u)
	}
	uRows.Close()

	query := "SELECT id, user_id, start_time, end_time, comment FROM slots"
	var args []interface{}
	if !startDate.IsZero() && !endDate.IsZero() {
		query += " WHERE start_time < ? AND end_time > ?"
		args = append(args, endDate, startDate)
	}
	
	sRows, _ := db.DB.Query(query, args...)
	for sRows.Next() {
		var s models.Slot
		sRows.Scan(&s.ID, &s.UserID, &s.StartTime, &s.EndTime, &s.Comment)
		data.Slots = append(data.Slots, s)
	}
	sRows.Close()

	eQuery := "SELECT id, title, description, icon, color, start_time, end_time FROM events"
	eRows, _ := db.DB.Query(eQuery)
	for eRows.Next() {
		var e EventRequest
		eRows.Scan(&e.ID, &e.Title, &e.Description, &e.Icon, &e.Color, &e.StartTime, &e.EndTime)
		
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
		data.Events = append(data.Events, e)
	}
	eRows.Close()

	if data.Users == nil { data.Users = []models.User{} }
	if data.Slots == nil { data.Slots = []models.Slot{} }
	if data.Events == nil { data.Events = []EventRequest{} }

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}