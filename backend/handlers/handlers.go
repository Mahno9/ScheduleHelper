package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"schedulehelper/backend/db"
	"schedulehelper/backend/models"
	"schedulehelper/backend/services"
)

// Handler holds all dependencies for HTTP handlers
type Handler struct {
	DB  *db.DB
	Hub *services.SSEHub
}

func New(database *db.DB, hub *services.SSEHub) *Handler {
	return &Handler{DB: database, Hub: hub}
}

// ---- HELPERS ----

func writeJSON(w http.ResponseWriter, code int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, code int, msg string) {
	writeJSON(w, code, map[string]string{"error": msg})
}

func parseID(r *http.Request, key string) (int64, error) {
	s := mux.Vars(r)[key]
	return strconv.ParseInt(s, 10, 64)
}

// parseTimeQuery parses an optional query param as RFC3339 time
func parseTimeQuery(r *http.Request, key string) time.Time {
	s := r.URL.Query().Get(key)
	if s == "" {
		return time.Time{}
	}
	t, _ := time.Parse(time.RFC3339, s)
	return t
}

// ---- HEALTH ----

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// ---- USERS ----

func (h *Handler) GetUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.DB.GetAllUsers()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get users")
		return
	}
	writeJSON(w, http.StatusOK, users)
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req models.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}
	if req.Emoji == "" {
		req.Emoji = "😊"
	}
	if req.Color == "" {
		req.Color = "#4A90D9"
	}

	// Check uniqueness
	existing, err := h.DB.GetUserByName(req.Name)
	if err == nil && existing != nil {
		writeError(w, http.StatusConflict, "name already taken")
		return
	}

	user, err := h.DB.CreateUser(req)
	if err != nil {
		log.Printf("register: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to create user")
		return
	}

	h.Hub.Broadcast("user_created", user)
	writeJSON(w, http.StatusCreated, user)
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	user, err := h.DB.GetUserByID(req.UserID)
	if err != nil {
		if err == sql.ErrNoRows {
			writeError(w, http.StatusNotFound, "user not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to get user")
		return
	}
	writeJSON(w, http.StatusOK, user)
}

func (h *Handler) GetProfile(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid user id")
		return
	}
	user, err := h.DB.GetUserByID(id)
	if err != nil {
		if err == sql.ErrNoRows {
			writeError(w, http.StatusNotFound, "user not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to get user")
		return
	}
	writeJSON(w, http.StatusOK, user)
}

func (h *Handler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid user id")
		return
	}
	var req models.UpdateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}
	if req.Theme == "" {
		req.Theme = "system"
	}

	// Check name uniqueness (allow keeping same name)
	existing, err := h.DB.GetUserByName(req.Name)
	if err == nil && existing != nil && existing.ID != id {
		writeError(w, http.StatusConflict, "name already taken")
		return
	}

	user, err := h.DB.UpdateUser(id, req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to update user")
		return
	}
	h.Hub.Broadcast("user_updated", user)
	writeJSON(w, http.StatusOK, user)
}

func (h *Handler) DeleteProfile(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid user id")
		return
	}
	if err := h.DB.DeleteUser(id); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to delete user")
		return
	}
	h.Hub.Broadcast("user_deleted", map[string]int64{"id": id})
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) UpdateSortOrder(w http.ResponseWriter, r *http.Request) {
	var req models.UpdateSortOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := h.DB.UpdateSortOrder(req.Order); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to update sort order")
		return
	}
	h.Hub.Broadcast("sort_order_updated", req)
	w.WriteHeader(http.StatusNoContent)
}

// ---- SLOTS ----

func (h *Handler) GetSlots(w http.ResponseWriter, r *http.Request) {
	var userID int64
	if uid := r.URL.Query().Get("user_id"); uid != "" {
		userID, _ = strconv.ParseInt(uid, 10, 64)
	}
	from := parseTimeQuery(r, "from")
	to := parseTimeQuery(r, "to")

	slots, err := h.DB.GetSlots(userID, from, to)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get slots")
		return
	}
	writeJSON(w, http.StatusOK, slots)
}

func (h *Handler) CreateSlot(w http.ResponseWriter, r *http.Request) {
	var body struct {
		UserID    int64  `json:"user_id"`
		StartTime string `json:"start_time"`
		EndTime   string `json:"end_time"`
		Comment   string `json:"comment"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	start, err := time.Parse(time.RFC3339, body.StartTime)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid start_time (use RFC3339)")
		return
	}
	end, err := time.Parse(time.RFC3339, body.EndTime)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid end_time (use RFC3339)")
		return
	}
	if !end.After(start) {
		writeError(w, http.StatusBadRequest, "end_time must be after start_time")
		return
	}
	slot, err := h.DB.CreateSlot(body.UserID, start, end, body.Comment)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create slot")
		return
	}
	h.Hub.Broadcast("slot_created", slot)
	writeJSON(w, http.StatusCreated, slot)
}

func (h *Handler) UpdateSlot(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid slot id")
		return
	}
	var body struct {
		StartTime string `json:"start_time"`
		EndTime   string `json:"end_time"`
		Comment   string `json:"comment"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	start, err := time.Parse(time.RFC3339, body.StartTime)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid start_time")
		return
	}
	end, err := time.Parse(time.RFC3339, body.EndTime)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid end_time")
		return
	}
	slot, err := h.DB.UpdateSlot(id, start, end, body.Comment)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to update slot")
		return
	}
	h.Hub.Broadcast("slot_updated", slot)
	writeJSON(w, http.StatusOK, slot)
}

func (h *Handler) DeleteSlot(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid slot id")
		return
	}
	if err := h.DB.DeleteSlot(id); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to delete slot")
		return
	}
	h.Hub.Broadcast("slot_deleted", map[string]int64{"id": id})
	w.WriteHeader(http.StatusNoContent)
}

// ---- EVENTS ----

func (h *Handler) GetEvents(w http.ResponseWriter, r *http.Request) {
	from := parseTimeQuery(r, "from")
	to := parseTimeQuery(r, "to")
	events, err := h.DB.GetEvents(from, to)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get events")
		return
	}
	writeJSON(w, http.StatusOK, events)
}

func (h *Handler) CreateEvent(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		Icon        string `json:"icon"`
		Color       string `json:"color"`
		StartTime   string `json:"start_time"`
		EndTime     string `json:"end_time"`
		CreatedBy   int64  `json:"created_by"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	body.Title = strings.TrimSpace(body.Title)
	if body.Title == "" {
		writeError(w, http.StatusBadRequest, "title is required")
		return
	}
	if len([]rune(body.Title)) > 120 {
		body.Title = string([]rune(body.Title)[:120])
	}
	if body.Color == "" {
		body.Color = "#D4AF37"
	}
	start, err := time.Parse(time.RFC3339, body.StartTime)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid start_time")
		return
	}
	end, err := time.Parse(time.RFC3339, body.EndTime)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid end_time")
		return
	}
	event, err := h.DB.CreateEvent(models.Event{
		Title: body.Title, Description: body.Description,
		Icon: body.Icon, Color: body.Color,
		StartTime: start, EndTime: end, CreatedBy: body.CreatedBy,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create event")
		return
	}
	h.Hub.Broadcast("event_created", event)
	writeJSON(w, http.StatusCreated, event)
}

func (h *Handler) UpdateEvent(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid event id")
		return
	}
	var body struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		Icon        string `json:"icon"`
		Color       string `json:"color"`
		StartTime   string `json:"start_time"`
		EndTime     string `json:"end_time"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if len([]rune(body.Title)) > 120 {
		body.Title = string([]rune(body.Title)[:120])
	}
	start, err := time.Parse(time.RFC3339, body.StartTime)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid start_time")
		return
	}
	end, err := time.Parse(time.RFC3339, body.EndTime)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid end_time")
		return
	}
	event, err := h.DB.UpdateEvent(id, models.Event{
		Title: body.Title, Description: body.Description,
		Icon: body.Icon, Color: body.Color,
		StartTime: start, EndTime: end,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to update event")
		return
	}
	h.Hub.Broadcast("event_updated", event)
	writeJSON(w, http.StatusOK, event)
}

func (h *Handler) DeleteEvent(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid event id")
		return
	}
	if err := h.DB.DeleteEvent(id); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to delete event")
		return
	}
	h.Hub.Broadcast("event_deleted", map[string]int64{"id": id})
	w.WriteHeader(http.StatusNoContent)
}

// ---- CALENDAR ----

func (h *Handler) GetCalendar(w http.ResponseWriter, r *http.Request) {
	from := parseTimeQuery(r, "from")
	to := parseTimeQuery(r, "to")
	if from.IsZero() {
		from = time.Now().Truncate(24 * time.Hour)
	}
	if to.IsZero() {
		to = from.Add(14 * 24 * time.Hour)
	}

	users, err := h.DB.GetAllUsers()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get users")
		return
	}
	slots, err := h.DB.GetSlots(0, from, to)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get slots")
		return
	}
	events, err := h.DB.GetEvents(from, to)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get events")
		return
	}

	// Filter users to only those with slots in period
	activeUIDs := make(map[int64]bool)
	for _, s := range slots {
		activeUIDs[s.UserID] = true
	}
	var activeUsers []models.User
	for _, u := range users {
		if activeUIDs[u.ID] {
			activeUsers = append(activeUsers, u)
		}
	}
	if activeUsers == nil {
		activeUsers = []models.User{}
	}

	writeJSON(w, http.StatusOK, models.CalendarData{
		Users:  activeUsers,
		Slots:  slots,
		Events: events,
	})
}

// ---- SSE ----

func (h *Handler) SSEEvents(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		writeError(w, http.StatusInternalServerError, "streaming not supported")
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	ch := h.Hub.Subscribe()
	defer h.Hub.Unsubscribe(ch)

	// Send initial ping
	fmt.Fprintf(w, "event: ping\ndata: {}\n\n")
	flusher.Flush()

	for {
		select {
		case <-r.Context().Done():
			return
		case data, ok := <-ch:
			if !ok {
				return
			}
			fmt.Fprintf(w, "data: %s\n\n", data)
			flusher.Flush()
		}
	}
}
