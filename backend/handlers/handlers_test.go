package handlers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"

	"github.com/gorilla/mux"
	"schedulehelper/backend/db"
	"schedulehelper/backend/handlers"
	"schedulehelper/backend/models"
	"schedulehelper/backend/services"
)

func setupTestEnv(t *testing.T) (*mux.Router, *handlers.Handler) {
	t.Helper()
	f, err := os.CreateTemp("", "schedulehelper-handler-test-*.db")
	if err != nil {
		t.Fatal(err)
	}
	f.Close()
	t.Cleanup(func() { os.Remove(f.Name()) })

	d, err := db.Open(f.Name())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { d.Close() })

	hub := services.NewSSEHub()
	h := handlers.New(d, hub)

	r := mux.NewRouter()
	api := r.PathPrefix("/api").Subrouter()
	api.HandleFunc("/health", h.Health).Methods(http.MethodGet)
	api.HandleFunc("/users", h.GetUsers).Methods(http.MethodGet)
	api.HandleFunc("/users/sort", h.UpdateSortOrder).Methods(http.MethodPut)
	api.HandleFunc("/register", h.Register).Methods(http.MethodPost)
	api.HandleFunc("/login", h.Login).Methods(http.MethodPost)
	api.HandleFunc("/profile/{id:[0-9]+}", h.GetProfile).Methods(http.MethodGet)
	api.HandleFunc("/profile/{id:[0-9]+}", h.UpdateProfile).Methods(http.MethodPut)
	api.HandleFunc("/profile/{id:[0-9]+}", h.DeleteProfile).Methods(http.MethodDelete)
	api.HandleFunc("/slots", h.GetSlots).Methods(http.MethodGet)
	api.HandleFunc("/slots", h.CreateSlot).Methods(http.MethodPost)
	api.HandleFunc("/slots/{id:[0-9]+}", h.UpdateSlot).Methods(http.MethodPut)
	api.HandleFunc("/slots/{id:[0-9]+}", h.DeleteSlot).Methods(http.MethodDelete)
	api.HandleFunc("/events", h.GetEvents).Methods(http.MethodGet)
	api.HandleFunc("/events", h.CreateEvent).Methods(http.MethodPost)
	api.HandleFunc("/events/{id:[0-9]+}", h.UpdateEvent).Methods(http.MethodPut)
	api.HandleFunc("/events/{id:[0-9]+}", h.DeleteEvent).Methods(http.MethodDelete)
	api.HandleFunc("/calendar", h.GetCalendar).Methods(http.MethodGet)
	return r, h
}

func doRequest(t *testing.T, r *mux.Router, method, path string, body interface{}) *httptest.ResponseRecorder {
	t.Helper()
	var buf bytes.Buffer
	if body != nil {
		json.NewEncoder(&buf).Encode(body)
	}
	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	return rr
}

func TestHealth(t *testing.T) {
	r, _ := setupTestEnv(t)
	rr := doRequest(t, r, "GET", "/api/health", nil)
	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
}

func TestRegisterAndLogin(t *testing.T) {
	r, _ := setupTestEnv(t)

	// Register
	rr := doRequest(t, r, "POST", "/api/register", map[string]interface{}{
		"name":  "TestUser",
		"emoji": "🤖",
		"color": "#123456",
	})
	if rr.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d: %s", rr.Code, rr.Body)
	}

	var user models.User
	json.NewDecoder(rr.Body).Decode(&user)
	if user.ID == 0 {
		t.Fatal("expected non-zero user ID")
	}

	// Login
	rr = doRequest(t, r, "POST", "/api/login", map[string]interface{}{
		"user_id": user.ID,
	})
	if rr.Code != http.StatusOK {
		t.Errorf("expected 200 login, got %d", rr.Code)
	}
}

func TestRegisterDuplicateName(t *testing.T) {
	r, _ := setupTestEnv(t)
	body := map[string]interface{}{"name": "DupUser", "emoji": "😀", "color": "#111"}
	rr := doRequest(t, r, "POST", "/api/register", body)
	if rr.Code != http.StatusCreated {
		t.Errorf("first register: expected 201, got %d", rr.Code)
	}
	rr = doRequest(t, r, "POST", "/api/register", body)
	if rr.Code != http.StatusConflict {
		t.Errorf("duplicate register: expected 409, got %d", rr.Code)
	}
}

func TestRegisterEmptyName(t *testing.T) {
	r, _ := setupTestEnv(t)
	rr := doRequest(t, r, "POST", "/api/register", map[string]interface{}{
		"name": "", "emoji": "😀", "color": "#000",
	})
	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for empty name, got %d", rr.Code)
	}
}

func TestGetUsersEmpty(t *testing.T) {
	r, _ := setupTestEnv(t)
	rr := doRequest(t, r, "GET", "/api/users", nil)
	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
	var users []models.User
	json.NewDecoder(rr.Body).Decode(&users)
	if len(users) != 0 {
		t.Errorf("expected empty users list, got %d", len(users))
	}
}

func TestGetProfile(t *testing.T) {
	r, _ := setupTestEnv(t)
	rr := doRequest(t, r, "POST", "/api/register", map[string]interface{}{
		"name": "ProfileUser", "emoji": "🎭", "color": "#555",
	})
	var user models.User
	json.NewDecoder(rr.Body).Decode(&user)

	rr = doRequest(t, r, "GET", "/api/profile/"+itoa(user.ID), nil)
	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
}

func TestUpdateProfile(t *testing.T) {
	r, _ := setupTestEnv(t)
	rr := doRequest(t, r, "POST", "/api/register", map[string]interface{}{
		"name": "OldName", "emoji": "🦊", "color": "#000",
	})
	var user models.User
	json.NewDecoder(rr.Body).Decode(&user)

	rr = doRequest(t, r, "PUT", "/api/profile/"+itoa(user.ID), map[string]interface{}{
		"name":          "NewName",
		"emoji":         "🦁",
		"color":         "#fff",
		"theme":         "dark",
		"timezone":      "Europe/Moscow",
		"auto_timezone": false,
	})
	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", rr.Code, rr.Body)
	}
	var updated models.User
	json.NewDecoder(rr.Body).Decode(&updated)
	if updated.Name != "NewName" {
		t.Errorf("expected name NewName, got %s", updated.Name)
	}
}

func TestDeleteProfile(t *testing.T) {
	r, _ := setupTestEnv(t)
	rr := doRequest(t, r, "POST", "/api/register", map[string]interface{}{
		"name": "ToDelete", "emoji": "🗑️", "color": "#000",
	})
	var user models.User
	json.NewDecoder(rr.Body).Decode(&user)

	rr = doRequest(t, r, "DELETE", "/api/profile/"+itoa(user.ID), nil)
	if rr.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rr.Code)
	}

	rr = doRequest(t, r, "GET", "/api/profile/"+itoa(user.ID), nil)
	if rr.Code != http.StatusNotFound {
		t.Errorf("expected 404 after delete, got %d", rr.Code)
	}
}

func TestSlotAPI(t *testing.T) {
	r, _ := setupTestEnv(t)
	rr := doRequest(t, r, "POST", "/api/register", map[string]interface{}{
		"name": "SlotAPIUser", "emoji": "⏰", "color": "#abc",
	})
	var user models.User
	json.NewDecoder(rr.Body).Decode(&user)

	// Create slot
	rr = doRequest(t, r, "POST", "/api/slots", map[string]interface{}{
		"user_id":    user.ID,
		"start_time": "2026-06-01T09:00:00Z",
		"end_time":   "2026-06-01T11:00:00Z",
		"comment":    "free slot",
	})
	if rr.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d: %s", rr.Code, rr.Body)
	}
	var slot models.Slot
	json.NewDecoder(rr.Body).Decode(&slot)
	if slot.ID == 0 {
		t.Fatal("expected non-zero slot ID")
	}

	// Get slots
	rr = doRequest(t, r, "GET", "/api/slots?user_id="+itoa(user.ID), nil)
	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
	var slots []models.Slot
	json.NewDecoder(rr.Body).Decode(&slots)
	if len(slots) != 1 {
		t.Errorf("expected 1 slot, got %d", len(slots))
	}

	// Update slot
	rr = doRequest(t, r, "PUT", "/api/slots/"+itoa(slot.ID), map[string]interface{}{
		"start_time": "2026-06-01T10:00:00Z",
		"end_time":   "2026-06-01T12:00:00Z",
		"comment":    "updated",
	})
	if rr.Code != http.StatusOK {
		t.Errorf("expected 200 on update, got %d", rr.Code)
	}

	// Delete slot
	rr = doRequest(t, r, "DELETE", "/api/slots/"+itoa(slot.ID), nil)
	if rr.Code != http.StatusNoContent {
		t.Errorf("expected 204 on delete, got %d", rr.Code)
	}
}

func TestSlotInvalidTime(t *testing.T) {
	r, _ := setupTestEnv(t)
	rr := doRequest(t, r, "POST", "/api/slots", map[string]interface{}{
		"user_id":    1,
		"start_time": "2026-06-01T11:00:00Z",
		"end_time":   "2026-06-01T09:00:00Z", // end before start
		"comment":    "",
	})
	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for invalid time range, got %d", rr.Code)
	}
}

func TestEventAPI(t *testing.T) {
	r, _ := setupTestEnv(t)
	rr := doRequest(t, r, "POST", "/api/register", map[string]interface{}{
		"name": "EventAPIUser", "emoji": "🎉", "color": "#gold",
	})
	var user models.User
	json.NewDecoder(rr.Body).Decode(&user)

	// Create event
	rr = doRequest(t, r, "POST", "/api/events", map[string]interface{}{
		"title":       "Stand-up",
		"description": "Daily stand-up call",
		"icon":        "📊",
		"color":       "#D4AF37",
		"start_time":  "2026-06-01T09:00:00Z",
		"end_time":    "2026-06-01T09:30:00Z",
		"created_by":  user.ID,
	})
	if rr.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d: %s", rr.Code, rr.Body)
	}
	var event models.Event
	json.NewDecoder(rr.Body).Decode(&event)

	// Get events
	rr = doRequest(t, r, "GET", "/api/events", nil)
	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
	var events []models.Event
	json.NewDecoder(rr.Body).Decode(&events)
	if len(events) != 1 {
		t.Errorf("expected 1 event, got %d", len(events))
	}

	// Update event
	rr = doRequest(t, r, "PUT", "/api/events/"+itoa(event.ID), map[string]interface{}{
		"title":       "Updated Stand-up",
		"description": "Updated",
		"icon":        "✅",
		"color":       "#D4AF37",
		"start_time":  "2026-06-01T09:00:00Z",
		"end_time":    "2026-06-01T09:30:00Z",
	})
	if rr.Code != http.StatusOK {
		t.Errorf("expected 200 on update, got %d", rr.Code)
	}

	// Delete event
	rr = doRequest(t, r, "DELETE", "/api/events/"+itoa(event.ID), nil)
	if rr.Code != http.StatusNoContent {
		t.Errorf("expected 204 on delete, got %d", rr.Code)
	}
}

func TestCalendar(t *testing.T) {
	r, _ := setupTestEnv(t)
	rr := doRequest(t, r, "GET", "/api/calendar", nil)
	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", rr.Code, rr.Body)
	}
}

func TestSortOrder(t *testing.T) {
	r, _ := setupTestEnv(t)
	var ids []int64
	for _, name := range []string{"Alpha", "Beta", "Gamma"} {
		rr := doRequest(t, r, "POST", "/api/register", map[string]interface{}{
			"name": name, "emoji": "😀", "color": "#000",
		})
		var u models.User
		json.NewDecoder(rr.Body).Decode(&u)
		ids = append(ids, u.ID)
	}
	rr := doRequest(t, r, "PUT", "/api/users/sort", map[string]interface{}{
		"order": []int64{ids[2], ids[0], ids[1]},
	})
	if rr.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d: %s", rr.Code, rr.Body)
	}
}

func itoa(id int64) string {
	return strconv.FormatInt(id, 10)
}
